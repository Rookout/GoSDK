package common

import (
	"fmt"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/sirupsen/logrus"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
)


const (
	GoogleProtoTypeURLPrefix          = "type.googleapis.com/"
	MessageTypeInitAugs               = "com.rookout.InitialAugsCommand"
	MessageTypeDataOnPremTokenRequest = "com.rookout.DataOnPremTokenRequest"
	MessageTypeInitC2C                = "com.rookout.C2CInitMessage"
	MessageTypeNewAgentMessage        = "com.rookout.NewAgentMessage"
	MessageTypeAddAugCommand          = "com.rookout.AddAugCommand"
	MessageTypeRemoveAugCommand       = "com.rookout.RemoveAugCommand"
	MessageTypeAgentInformation       = "com.rookout.AgentInformation"
	MessageTypeAgentInformationArray  = "com.rookout.AgentInformationArray"
	MessageTypeLogMessage             = "com.rookout.LogMessage" 
	MessageTypeControllerLogMessage   = "com.rookout.ControllerLogMessage"
	MessageTypeRuleStatusMessage      = "com.rookout.RuleStatusMessage"
	MessageTypeAugReport              = "com.rookout.AugReportMessage"
	MessageTypeUserMsg                = "com.rookout.UserMsg"
	MessageTypePingMessage            = "com.rookout.PingMessage"
	MessageTypeHitCountUpdateMessage  = "com.rookout.HitCountUpdateMessage"
	MessageTypeAgentsListMessage      = "com.rookout.AgentsList"
	MessageTypeRpcRequest             = "com.rookout.C2cRpcRequest"
	MessageTypeRpcResponse            = "com.rookout.C2cRpcResponse"
	MessageTypeAssertAgents           = "com.rookout.AssertAgents"
)

func WrapMsgInEnvelopeWithTime(message proto.Message, t time.Time) ([]byte, error) {
	envelope := &pb.Envelope{}
	msgAsAny, err := AnyProtoMarshal(message)
	envelope.Msg = msgAsAny
	envelope.Timestamp, _ = ptypes.TimestampProto(t)

	if err != nil {
		return nil, err
	}

	payload, err := proto.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal envelope (%s) message: %s", msgAsAny.TypeUrl, err.Error())
	}
	return payload, nil
}


func WrapMsgInEnvelope(message proto.Message) ([]byte, error) {
	return WrapMsgInEnvelopeWithTime(message, time.Now())
}


func ParseEnvelope(buf []byte) (*pb.Envelope, string, error) {
	envelope := &pb.Envelope{}

	err := proto.Unmarshal(buf, envelope)
	if err != nil {
		return nil, "", err
	}

	typeName := NormalizeType(envelope.Msg.TypeUrl)

	return envelope, typeName, nil
}

func NormalizeType(typeName string) string {
	return strings.Replace(typeName, GoogleProtoTypeURLPrefix, "", 1)
}

func AnyProtoMarshal(message proto.Message) (*any.Any, error) {
	msgAsAny, err := ptypes.MarshalAny(message)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal message err - %s", err.Error())
	}

	return msgAsAny, nil
}

func AnyProtoUnmarshal(anyMsg *any.Any, msg proto.Message, msgType string) error {
	err := ptypes.UnmarshalAny(anyMsg, msg)

	if err != nil {
		logrus.WithError(err).Errorf("Failed to convert any.Any type message to: %s msg: %s", msgType, TruncateString(msg.String(), 50))
		return err
	}

	return nil
}

func TruncateString(str string, num int) string {
	if num > len(str) {
		return str
	}
	return str[:num]
}

func ExtractProtoFromEnvelope(envelope *pb.Envelope, msg proto.Message, msgType string) error {
	err := ptypes.UnmarshalAny(envelope.Msg, msg)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to convert envelope message to: %s, msg: %s", msgType, TruncateString(msg.String(), 50))
		return err
	}
	return nil
}
