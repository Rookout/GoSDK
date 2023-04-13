package com_ws

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Rookout/GoSDK/pkg/common"
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/logger"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/utils"
	"github.com/go-errors/errors"
	gorilla "github.com/gorilla/websocket"
)

var dialer *gorilla.Dialer
var dialerOnce sync.Once

type WebSocketClientCreator func(context.Context, *url.URL, string, *url.URL, *pb.AgentInformation) WebSocketClient

type WebSocketClient interface {
	GetConnectionCtx() context.Context
	Dial(context.Context) error
	Handshake(context.Context) error
	Receive(context.Context) ([]byte, error)
	Send(context.Context, []byte) error
	Close()
}

type webSocketClient struct {
	agentURL            *url.URL
	agentInfo           *pb.AgentInformation
	conn                *gorilla.Conn
	token               string
	proxy               *url.URL
	ConnectionCtx       context.Context
	cancelConnectionCtx context.CancelFunc
	writeMutex          sync.Mutex
}

func NewWebSocketClient(ctx context.Context, agentURL *url.URL, token string, proxy *url.URL, agentInfo *pb.AgentInformation) WebSocketClient {
	client := &webSocketClient{
		agentURL:  agentURL,
		agentInfo: agentInfo,
		conn:      &gorilla.Conn{},
		token:     token,
		proxy:     proxy,
	}
	client.ConnectionCtx, client.cancelConnectionCtx = context.WithCancel(ctx)
	return client
}

func (w *webSocketClient) GetConnectionCtx() context.Context {
	return w.ConnectionCtx
}

func (w *webSocketClient) Dial(ctx context.Context) error {
	conn, httpRes, err := w.getWSDialer().DialContext(ctx, w.agentURL.String(), http.Header{"X-Rookout-Token": []string{w.token}})
	if err != nil {
		badToken := isHttpResponseBadToken(httpRes)
		if badToken {
			censoredToken := ""
			if len(w.token) > 5 {
				censoredToken = w.token[:5]
			}

			logger.Logger().Errorf("The Rookout token supplied (%s) is not valid; please check the token and try again", censoredToken)
			return rookoutErrors.NewInvalidTokenError()
		} else {
			logger.Logger().Errorf("Failed to connect to controller (%s). err: %s", w.agentURL, err.Error())
		}
		return err
	}
	w.conn = conn

	pingTimeout := config.WebSocketClientConfig().PingTimeout
	if err = w.conn.SetReadDeadline(time.Now().Add(pingTimeout)); err != nil {
		logger.Logger().WithError(err).Error("failed to set read deadline, closing connection")
		w.Close()
		return err
	}
	utils.CreateGoroutine(func() {
		w.sendPingLoop()
	})
	w.conn.SetPongHandler(func(string) error {
		err := w.conn.SetReadDeadline(time.Now().Add(pingTimeout))
		if err != nil {
			logger.Logger().WithError(err).Error("Failed to set read deadline on pong, closing connection")
			w.Close()
		}

		return nil
	})

	return nil
}

func (w *webSocketClient) Handshake(ctx context.Context) error {
	buf, err := common.WrapMsgInEnvelope(&pb.NewAgentMessage{AgentInfo: w.agentInfo})
	if err != nil {
		return err
	}

	err = w.Send(ctx, buf)
	if err != nil {
		return err
	}

	return nil
}

func (w *webSocketClient) Receive(ctx context.Context) ([]byte, error) {
	
	if deadline, ok := ctx.Deadline(); ok {
		err := w.conn.SetReadDeadline(deadline)
		if err != nil {
			return nil, err
		}
	}
	messageType, buf, err := w.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	if messageType != gorilla.BinaryMessage {
		return nil, errors.Errorf("unexpected message type, got %d\n", messageType)
	}

	return buf, nil
}

func (w *webSocketClient) sendPing(ctx context.Context) error {
	err := w.sendMsg(ctx, gorilla.PingMessage, nil)
	if err != nil {
		return err
	}
	return nil
}

func (w *webSocketClient) sendPingLoop() {
	defer w.cancelConnectionCtx()

	pingTimer := time.NewTicker(config.WebSocketClientConfig().PingInterval)
	defer drainTimer(pingTimer)
	defer pingTimer.Stop()

	for {
		select {
		case <-w.ConnectionCtx.Done():
			return
		case <-pingTimer.C:
			err := func() error {
				ctxTimeout, cancelFunc := context.WithTimeout(w.ConnectionCtx, config.WebSocketClientConfig().WriteTimeout)
				defer cancelFunc()

				return w.sendPing(ctxTimeout)
			}()
			if err != nil {
				logger.Logger().WithError(err).Error("Failed writing ping")
				return
			}
		}
	}
}

func (w *webSocketClient) sendMsg(ctx context.Context, msgType int, data []byte) error {
	w.writeMutex.Lock()
	defer w.writeMutex.Unlock()

	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		err := w.conn.SetWriteDeadline(deadline)
		if err != nil {
			return err
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return w.conn.WriteMessage(msgType, data)
}

func (w *webSocketClient) sendBinary(ctx context.Context, buf []byte) error {
	err := w.sendMsg(ctx, gorilla.BinaryMessage, buf)
	if err != nil {
		return err
	}
	return nil
}

func (w *webSocketClient) Send(ctx context.Context, buf []byte) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	err := func() error {
		ctxTimeout, cancelFunc := context.WithTimeout(ctx, config.WebSocketClientConfig().WriteTimeout)
		defer cancelFunc()

		return w.sendBinary(ctxTimeout, buf)
	}()
	if err != nil {
		logger.Logger().WithError(err).Error("Failed writing message")
		return err
	}
	return nil
}

func (w *webSocketClient) Close() {
	_ = w.conn.Close()
	w.cancelConnectionCtx()
}

func isHttpResponseBadToken(httpRes *http.Response) bool {
	if httpRes == nil {
		return false
	}
	return httpRes.StatusCode == http.StatusForbidden || httpRes.StatusCode == http.StatusUnauthorized
}

func drainTimer(timer *time.Ticker) {
	select {
	case <-timer.C:
	default:
	}
}

func (w *webSocketClient) getWSDialer() *gorilla.Dialer {
	dialerOnce.Do(func() {
		dialerTemp := *gorilla.DefaultDialer
		netDialer := net.Dialer{Resolver: &net.Resolver{PreferGo: true}}
		dialerTemp.NetDial = netDialer.Dial
		dialer = &dialerTemp
		dialerTemp.TLSClientConfig = &tls.Config{InsecureSkipVerify: config.WebSocketClientConfig().SkipSSLVerify}
	})

	if w.proxy != nil {
		dialer.Proxy = func(_ *http.Request) (*url.URL, error) {
			return w.proxy, nil
		}
		logger.Logger().Infof("Using proxy: %s", w.proxy.String())
	}
	return dialer
}
