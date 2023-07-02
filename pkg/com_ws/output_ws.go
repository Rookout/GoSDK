package com_ws

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Rookout/GoSDK/pkg/common"
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/processor/namespaces"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/Rookout/GoSDK/pkg/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Output interface {
	SetAgentID(agentID string)

	SendUserMessage(augID types.AugID, messageID string, arguments namespaces.Namespace)
	SendRuleStatus(augID types.AugID, active string, errIn rookoutErrors.RookoutError) error
	SendWarning(augID types.AugID, err rookoutErrors.RookoutError) error
	SendError(augID types.AugID, err rookoutErrors.RookoutError) error
	SendOutputQueueFullWarning(augID types.AugID)

	SetAgentCom(agentCom AgentCom)

	StopSendingMessages()
}

type outputWs struct {
	agentID       string
	agentCom      AgentCom
	closed        atomic.Value
	skippedAugIDs *SyncSet
}

func NewOutputWs() *outputWs {
	o := &outputWs{
		closed:        atomic.Value{},
		skippedAugIDs: newSyncSet(),
	}
	o.closed.Store(false)
	return o
}

func (d *outputWs) SetAgentCom(agentCom AgentCom) {
	d.agentCom = agentCom
}

func (d *outputWs) SetAgentID(agentID string) {
	d.agentID = agentID
}

func (d *outputWs) SendUserMessage(augID types.AugID, messageID string, arguments namespaces.Namespace) {
	utils.CreateGoroutine(func() {
		defer func() {
			if closer, ok := arguments.(io.Closer); ok {
				_ = closer.Close()
			}
		}()

		err := d.sendUserMessage(augID, messageID, arguments)
		if err != nil {
			logger.Logger().WithError(err).Warningf("Unable to send user message, aug id: %s", augID)
		}
	})
}

func (d *outputWs) sendUserMessage(augID types.AugID, messageID string, arguments namespaces.Namespace) error {
	if d.closed.Load().(bool) {
		return nil
	}

	fmt.Printf("Sending augID: %d, messageID: %d\n", augID, messageID)
	msg := &pb.AugReportMessage{
		AgentId:  d.agentID,
		AugID:    augID,
		ReportID: messageID,
	}
	if config.OutputWsConfig().ProtobufVersion2 {
		serializer := namespaces.NewNamespaceSerializer2(arguments, true)
		msg.Arguments2 = serializer.Variant2
		msg.StringsCache = serializer.StringCache
	} else {
		serializer := namespaces.NewNamespaceSerializer(arguments, true)
		msg.Arguments = serializer.Variant
	}
	buf, err := common.WrapMsgInEnvelope(msg)
	if err != nil {
		return err
	}

	rookoutErr := d.agentCom.Send(buf)
	if rookoutErr != nil && rookoutErr.GetType() == "RookOutputQueueFull" {
		d.SendOutputQueueFullWarning(augID)
		return rookoutErr
	}

	d.skippedAugIDs.Remove(augID)
	return nil
}

func (d *outputWs) SendRuleStatus(augID types.AugID, active string, errIn rookoutErrors.RookoutError) error {
	if d.closed.Load().(bool) {
		return nil
	}

	if active == "Deleted" {
		d.skippedAugIDs.Remove(augID)
	}

	stat := &pb.RuleStatusMessage{
		AgentId: d.agentID,
		RuleId:  augID,
		Active:  active,
	}
	if errIn != nil {
		serializer := namespaces.NewNamespaceSerializer(namespaces.NewGoObjectNamespace(errIn), true)
		stat.Error = serializer.GetErrorValue()
	}

	buf, err := common.WrapMsgInEnvelope(stat)
	if err != nil {
		return err
	}
	return d.agentCom.Send(buf)
}

func (d *outputWs) SendWarning(augID types.AugID, err rookoutErrors.RookoutError) error {
	return d.SendRuleStatus(augID, "Warning", err)
}

func (d *outputWs) SendError(augID types.AugID, err rookoutErrors.RookoutError) error {
	return d.SendRuleStatus(augID, "Error", err)
}

func (d *outputWs) SendLogMessage(level pb.LogMessage_LogLevel, time time.Time, filename string, lineno int, text string, arguments map[string]interface{}) error {
	if d.closed.Load().(bool) || d.agentCom == nil {
		return nil
	}

	argumentsNamespace := namespaces.NewEmptyContainerNamespace()
	parametersNamespace := namespaces.NewEmptyContainerNamespace()
	for k, v := range arguments {
		if v == nil {
			continue
		}

		if k == "error" {
			_ = argumentsNamespace.WriteAttribute("exc", namespaces.NewGoObjectNamespace(v))
		} else {
			_ = parametersNamespace.WriteAttribute(k, namespaces.NewGoObjectNamespace(v))
		}
	}
	_ = argumentsNamespace.WriteAttribute("args", parametersNamespace)
	serializer := namespaces.NewNamespaceSerializer(argumentsNamespace, false)

	timestamp := timestamppb.New(time)
	msg := &pb.LogMessage{
		Timestamp:        timestamp,
		AgentId:          d.agentID,
		Level:            level,
		Filename:         filename,
		Line:             uint32(lineno),
		Text:             text,
		FormattedMessage: text,
		Arguments:        []*pb.Variant{serializer.Variant},
		LegacyArguments:  serializer.Variant,
	}
	buf, err := common.WrapMsgInEnvelope(msg)
	if err != nil {
		return err
	}
	return d.agentCom.Send(buf)
}

func (d *outputWs) StopSendingMessages() {
	d.closed.Store(true)
}

func (d *outputWs) SendOutputQueueFullWarning(augID types.AugID) {
	if d.skippedAugIDs.Contains(augID) {
		return
	}

	d.skippedAugIDs.Add(augID)
	_ = d.SendRuleStatus(augID, "Warning", rookoutErrors.NewRookOutputQueueFull())
	logger.Logger().Warning("Skipping aug (" + augID + ") execution because the queue is full")
}

type SyncSet struct {
	internalMap map[types.AugID]struct{}
	lock        sync.Mutex
}

func newSyncSet() *SyncSet {
	return &SyncSet{internalMap: make(map[types.AugID]struct{})}
}

func (s *SyncSet) Add(value interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.internalMap[value.(types.AugID)] = struct{}{}
}

func (s *SyncSet) Remove(value interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.internalMap, value.(types.AugID))
}

func (s *SyncSet) Contains(value interface{}) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.internalMap[value.(types.AugID)]
	return ok
}
