package com_ws

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/Rookout/GoSDK/pkg/common"
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/information"
	"github.com/Rookout/GoSDK/pkg/logger"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/utils"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/anypb"
)

type Callable func(*anypb.Any)

type AgentCom interface {
	ConnectToAgent() error
	RegisterCallback(string, Callable)
	Send([]byte) rookoutErrors.RookoutError
	Stop()
	Flush()
}

type messageCallback struct {
	callback   Callable
	persistent bool
}

type agentComWs struct {
	agentID                  string
	output                   Output
	agentURL                 *url.URL
	proxy                    *url.URL
	token                    string
	callbacks                map[string][]messageCallback
	agentInfo                *pb.AgentInformation
	printOnInitialConnection bool
	stopCtx                  context.Context
	stopCtxCancel            context.CancelFunc
	outgoingChan             *SizeLimitedChannel
	gotInitialAugs           chan bool
	clientCreator            WebSocketClientCreator
	client                   WebSocketClient
	backoff                  Backoff
}

func NewAgentComWs(clientCreator WebSocketClientCreator, output Output, backoff Backoff, agentHost string, agentPort int, proxy string,
	token string, labels map[string]string, printOnInitialConnection bool) (*agentComWs, error) {
	var a agentComWs
	var err error
	a.stopCtx, a.stopCtxCancel = context.WithCancel(context.Background())
	a.setId()
	a.agentURL, err = buildAgentURL(agentHost, agentPort)
	if err != nil {
		return nil, err
	}
	proxyUrl, err := buildProxyURL(proxy)
	if err != nil {
		logger.Logger().Fatalln("Bad proxy address: " + err.Error())
		return nil, err
	}
	a.proxy = proxyUrl
	a.agentInfo, err = information.Collect(labels, "")
	if err != nil {
		return nil, err
	}
	a.agentInfo.AgentId = a.agentID
	a.token = token
	a.callbacks = map[string][]messageCallback{}
	a.printOnInitialConnection = printOnInitialConnection
	a.outgoingChan = NewSizeLimitedChannel()
	a.gotInitialAugs = make(chan bool, 1)
	a.clientCreator = clientCreator
	a.backoff = backoff
	a.output = output
	a.output.SetAgentID(a.agentID)

	return &a, nil
}

func buildProxyURL(proxy string) (*url.URL, error) {
	if proxy == "" {
		return nil, nil
	}
	if !strings.Contains(proxy, "://") {
		proxy = "http://" + proxy
	}
	return url.Parse(proxy)
}

func buildAgentURL(agentHost string, agentPort int) (*url.URL, error) {
	if agentHost != "" && !strings.Contains(agentHost, "://") {
		agentHost = "ws://" + agentHost
	}
	urlString := fmt.Sprintf("%s:%d/v1", agentHost, agentPort)
	return url.Parse(urlString)
}

func (a *agentComWs) setId() {
	id, _ := uuid.New().MarshalBinary()
	a.agentID = hex.EncodeToString(id)
}

func (a *agentComWs) on(messageName string, callback Callable, persistent bool) {
	messageCallback := messageCallback{callback, persistent}
	a.callbacks[messageName] = append(a.callbacks[messageName], messageCallback)
}

func (a *agentComWs) RegisterCallback(messageName string, callback Callable) {
	a.on(messageName, callback, true)
}

func (a *agentComWs) ConnectToAgent() error {
	connectionTimeoutCtx, cancelConnectionTimeoutCtx := context.WithTimeout(context.Background(), config.AgentComWsConfig().ConnectionTimeout)
	defer cancelConnectionTimeoutCtx()
	connErrorsChan := make(chan error)

	utils.CreateRetryingGoroutine(a.stopCtx, func() { a.connectLoop(connErrorsChan) })

	select {
	case <-connectionTimeoutCtx.Done():
		return rookoutErrors.NewRookConnectToControllerTimeout()
	case err := <-connErrorsChan:
		return err
	}
}

func (a *agentComWs) Stop() {
	a.output.StopSendingMessages()
	select {
	case <-a.stopCtx.Done():
	default:
		a.stopCtxCancel()
	}

	if a.client != nil {
		a.client.Close()
	}
}

func (a *agentComWs) Flush() {
	err := a.outgoingChan.Flush()
	if err != nil {
		logger.Logger().WithError(err).Info("Flush failed")
	}
}

func (a *agentComWs) connectLoop(connErrorsChan chan error) {
	for {
		if !a.isRunning() {
			return
		}

		logger.Logger().Info("Connecting to controller.")
		connectionCtx, err := func() (context.Context, error) {
			connectCtx, cancelConnectCtx := context.WithTimeout(a.stopCtx, config.AgentComWsConfig().ConnectTimeout)
			defer cancelConnectCtx()
			return a.connect(connectCtx)
		}()
		if err != nil {
			logger.Logger().WithError(err).Info("Failed to connect to controller")
			select {
			case connErrorsChan <- err:
			default:
			}
			a.backoff.AfterDisconnect(a.stopCtx)
			continue
		}

		a.backoff.AfterConnect()
		select {
		case connErrorsChan <- nil:
		default:
		}
		if a.printOnInitialConnection {
			a.printOnInitialConnection = false
			logger.QuietPrintln("[Rookout] Successfully connected to controller.")
			logger.Logger().Debug("[Rookout] Agent ID is " + a.agentID)
		}
		logger.Logger().Info("Connected successfully to cloud controller")
		logger.Logger().Info("Finished initialization")

		select {
		case <-a.stopCtx.Done():
			return
		case <-connectionCtx.Done():
			a.client.Close()
			logger.Logger().Info("Disconnected from controller")
			a.backoff.AfterDisconnect(a.stopCtx)
		}
	}
}

func (a *agentComWs) connect(ctx context.Context) (context.Context, error) {
	
	a.client = a.clientCreator(a.stopCtx, a.agentURL, a.token, a.proxy, a.agentInfo)
	err := a.dialAndHandshake(ctx, a.client)
	if err != nil {
		return nil, err
	}

	
	a.on(common.MessageTypeInitAugs, func(any *anypb.Any) {
		a.gotInitialAugs <- true
	}, false)

	connectionCtx, cancelConnectionCtx := context.WithCancel(a.client.GetConnectionCtx())
	utils.CreateGoroutine(func() {
		defer cancelConnectionCtx()
		a.sendLoop(connectionCtx, a.client)
	})
	utils.CreateGoroutine(func() {
		defer cancelConnectionCtx()
		a.receiveLoop(connectionCtx, a.client)
	})

	select {
	case <-a.gotInitialAugs:
		return connectionCtx, nil
	case <-ctx.Done():
		return nil, rookoutErrors.NewContextEnded(ctx.Err())
	case <-connectionCtx.Done():
		return nil, rookoutErrors.NewContextEnded(connectionCtx.Err())
	}
}

func (a *agentComWs) sendLoop(ctx context.Context, client WebSocketClient) {
	for {
		
		buf := a.outgoingChan.Poll(ctx)
		if buf == nil {
			return
		}
		err := client.Send(ctx, buf)
		if err != nil {
			logger.Logger().WithError(err).Error("Failed when sending a message")
			
			_ = a.outgoingChan.Offer(buf)
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (a *agentComWs) receiveLoop(ctx context.Context, client WebSocketClient) {
	for {
		
		buf, err := client.Receive(ctx)
		if err != nil {
			logger.Logger().WithError(err).Error("failed when receiving a message")
			return
		}

		envelope, typeName, err := common.ParseEnvelope(buf)
		if err != nil {
			logger.Logger().WithError(err).Infof("failed to parse message from controller")
			continue
		}
		a.handleIncomingMessage(typeName, envelope)

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (a *agentComWs) Send(buf []byte) rookoutErrors.RookoutError {
	return a.outgoingChan.Offer(buf)
}

func (a *agentComWs) isRunning() bool {
	select {
	case <-a.stopCtx.Done():
		return false
	default:
		return true
	}
}

func (a *agentComWs) dialAndHandshake(ctx context.Context, client WebSocketClient) error {
	logger.Logger().Info("Attempting connection to cloud controller")
	err := client.Dial(ctx)
	if err != nil {
		return err
	}
	logger.Logger().Info("Dial to cloud controller returned")

	logger.Logger().Info("Starting handshake with cloud controller")
	err = client.Handshake(ctx)
	if err != nil {
		logger.Logger().WithError(err).Error("websocket handshake failed")
		client.Close()
		return err
	}
	logger.Logger().Info("Handshake with cloud controller completed successfully")
	return nil
}

func (a *agentComWs) handleIncomingMessage(typeName string, envelope *pb.Envelope) {
	var persistentCallbacks []messageCallback
	if callbacks, exists := a.callbacks[typeName]; exists {
		for _, messageCB := range callbacks {
			messageCB.callback(envelope.GetMsg())

			if messageCB.persistent {
				persistentCallbacks = append(persistentCallbacks, messageCB)
			}
		}
		a.callbacks[typeName] = persistentCallbacks
	} else {
		logger.Logger().Infof("Received unknown command: %s", typeName)
	}
}
