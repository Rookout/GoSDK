package aug_manager

import (
	"encoding/json"
	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/common"
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/logger"
	pb "github.com/Rookout/GoSDK/pkg/protobuf"
	"github.com/Rookout/GoSDK/pkg/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type CommandHandler struct {
	agentCom   com_ws.AgentCom
	augManager AugManager
}

func NewCommandHandler(agentCom com_ws.AgentCom, augManager AugManager) *CommandHandler {
	handler := CommandHandler{agentCom, augManager}
	handler.agentCom.RegisterCallback(common.MessageTypeInitAugs, func(msg *anypb.Any) {
		initAugs := &pb.InitialAugsCommand{}
		err := anypb.UnmarshalTo(msg, initAugs, proto.UnmarshalOptions{})
		if err != nil {
			logger.Logger().WithError(err).Errorf("failed to unmarshal init augs")
			return
		}

		newConfig := config.ParseConfig(initAugs.SdkConfiguration)
		config.UpdateGlobalRateLimitConfig(&newConfig)
		handler.updateConfig(newConfig)

		augs, err := handler.buildAugMap(initAugs.Augs)
		if err != nil {
			logger.Logger().WithError(err).Errorf("failed to build rules")
			return
		}

		handler.augManager.InitializeAugs(augs)
	})

	handler.agentCom.RegisterCallback(common.MessageTypeRemoveAugCommand, func(msg *anypb.Any) {
		removeAugCmd := &pb.RemoveAugCommand{}
		err := anypb.UnmarshalTo(msg, removeAugCmd, proto.UnmarshalOptions{})
		if err != nil {
			logger.Logger().WithError(err).Error("Failed to unmarshal envelope to RemoveAugCommand")
			return
		}
		err = handler.augManager.RemoveAug(removeAugCmd.AugId)
		if err != nil {
			logger.Logger().WithError(err).Error("failed to remove rule")
		}
	})

	handler.agentCom.RegisterCallback(common.MessageTypeAddAugCommand, func(msg *anypb.Any) {
		addAugCmd := &pb.AddAugCommand{}
		err := anypb.UnmarshalTo(msg, addAugCmd, proto.UnmarshalOptions{})
		if err != nil {
			logger.Logger().WithError(err).Error("failed to unmarshal envelope to AddAugCommand")
			return
		}
		augConfig := make(types.AugConfiguration)
		err = json.Unmarshal([]byte(addAugCmd.AugJson), &augConfig)
		if err != nil {
			logger.Logger().WithError(err).Error("failed to parse Rule")
			return
		}
		handler.augManager.AddAug(augConfig)
	})

	return &handler
}

func (c *CommandHandler) updateConfig(newConfig config.DynamicConfiguration) {
	config.UpdateObjectDumpConfigDefaults(newConfig.ObjectDumpConfigDefaults)
	c.agentCom.UpdateConfig(newConfig.AgentComWsConfiguration)
	c.augManager.UpdateConfig(newConfig.LocationsConfiguration)
}

func (c *CommandHandler) buildAugMap(rules []string) (map[types.AugId]types.AugConfiguration, error) {
	rulesMap := make(map[types.AugId]types.AugConfiguration)

	for _, ruleStr := range rules {
		augConfig := make(types.AugConfiguration)
		err := json.Unmarshal([]byte(ruleStr), &augConfig)
		if err != nil {
			logger.Logger().WithError(err).Error("failed to parse aug")
			return nil, err
		}

		if augId, ok := augConfig["id"].(types.AugId); ok {
			rulesMap[augId] = augConfig
		}
	}
	return rulesMap, nil
}
