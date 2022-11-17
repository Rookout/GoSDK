package com_ws

import (
	"github.com/Rookout/GoSDK/pkg/common"
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/processor/namespaces"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/Rookout/GoSDK/pkg/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type Output interface {
	SetAgentId(agentId string)

	SendUserMessage(augId types.AugId, messageId string, arguments types.Namespace)
	SendRuleStatus(augId types.AugId, active string, errIn rookoutErrors.RookoutError) error
	SendWarning(augId types.AugId, err rookoutErrors.RookoutError) error
	SendError(augId types.AugId, err rookoutErrors.RookoutError) error
	SendOutputQueueFullWarning(augId types.AugId)

	SetAgentCom(agentCom AgentCom)

	StopSendingMessages()
}

type outputWs struct {
	config        config.OutputWsConfiguration
	agentId       string
	agentCom      AgentCom
	closed        atomic.Value
	skippedAugIds *SyncSet
}

func NewOutputWs(config config.OutputWsConfiguration) *outputWs {
	o := &outputWs{
		closed:        atomic.Value{},
		config:        config,
		skippedAugIds: newSyncSet(),
	}
	o.closed.Store(false)
	return o
}

func (d *outputWs) SetAgentCom(agentCom AgentCom) {
	d.agentCom = agentCom
}

func (d *outputWs) SetAgentId(agentId string) {
	d.agentId = agentId
}

func (d *outputWs) SendUserMessage(augId types.AugId, messageId string, arguments types.Namespace) {
	utils.CreateGoroutine(func() {
		defer func() {
			if closer, ok := arguments.(io.Closer); ok {
				_ = closer.Close()
			}
		}()

		err := d.sendUserMessage(augId, messageId, arguments)
		if err != nil {
			logger.Logger().WithError(err).Warningf("Unable to send user message, aug id: %s", augId)
		}
	})
}

func (d *outputWs) sendUserMessage(augId types.AugId, messageId string, arguments types.Namespace) error {
	if d.closed.Load().(bool) {
		return nil
	}

	msg := &pb.AugReportMessage{
		AgentId:   d.agentId,
		AugId:     augId,
		Arguments: arguments.ToProtobuf(true),
		ReportId:  messageId,
	}
	buf, err := common.WrapMsgInEnvelope(msg)
	if err != nil {
		return err
	}

	rookoutErr := d.agentCom.Send(buf)
	if rookoutErr != nil && rookoutErr.GetType() == "RookOutputQueueFull" {
		d.SendOutputQueueFullWarning(augId)
		return rookoutErr
	}

	d.skippedAugIds.Remove(augId)
	return nil
}

func (d *outputWs) SendRuleStatus(augId types.AugId, active string, errIn rookoutErrors.RookoutError) error {
	if d.closed.Load().(bool) {
		return nil
	}

	if active == "Deleted" {
		d.skippedAugIds.Remove(augId)
	}

	stat := &pb.RuleStatusMessage{
		AgentId: d.agentId,
		RuleId:  augId,
		Active:  active,
		Error:   namespaces.GetErrorAsProtobuf(errIn),
	}
	buf, err := common.WrapMsgInEnvelope(stat)
	if err != nil {
		return err
	}
	return d.agentCom.Send(buf)
}

func (d *outputWs) SendWarning(augId types.AugId, err rookoutErrors.RookoutError) error {
	return d.SendRuleStatus(augId, "Warning", err)
}

func (d *outputWs) SendError(augId types.AugId, err rookoutErrors.RookoutError) error {
	return d.SendRuleStatus(augId, "Error", err)
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

	timestamp := timestamppb.New(time)
	msg := &pb.LogMessage{
		Timestamp:        timestamp,
		AgentId:          d.agentId,
		Level:            level,
		Filename:         filename,
		Line:             uint32(lineno),
		Text:             text,
		FormattedMessage: text,
		Arguments:        []*pb.Variant{argumentsNamespace.ToProtobuf(false)},
		LegacyArguments:  argumentsNamespace.ToProtobuf(false),
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

func (d *outputWs) SendOutputQueueFullWarning(augId types.AugId) {
	if d.skippedAugIds.Contains(augId) {
		return
	}

	d.skippedAugIds.Add(augId)
	_ = d.SendRuleStatus(augId, "Warning", rookoutErrors.NewRookOutputQueueFull())
	logger.Logger().Warning("Skipping aug (" + augId + ") execution because the queue is full")
}

type SyncSet struct {
	internalMap sync.Map
}

func newSyncSet() *SyncSet {
	return &SyncSet{}
}

func (s *SyncSet) Add(value interface{}) {
	s.internalMap.Store(value, struct{}{})
}

func (s *SyncSet) Remove(value interface{}) {
	s.internalMap.Delete(value)
}

func (s *SyncSet) Contains(value interface{}) bool {
	_, ok := s.internalMap.Load(value)
	return ok
}
