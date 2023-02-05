package locations

import (
	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
)

type HashInfo struct {
}

type LocationFileLine struct {
	filename string
	lineno   int
	hashInfo *HashInfo
	output   com_ws.Output
	aug      augs.Aug

	status string
}

func (l *LocationFileLine) sendRuleStatus(status string, err rookoutErrors.RookoutError) error {
	if l.status == status {
		return nil
	}

	logger.Logger().WithError(err).Infof("Updating rule status for: %s to %s\n", l.GetAugId(), status)

	l.status = status
	return l.output.SendRuleStatus(l.GetAugId(), status, err)
}

func (l *LocationFileLine) SetPending() error {
	return l.sendRuleStatus("Pending", nil)
}

func (l *LocationFileLine) SetActive() error {
	return l.sendRuleStatus("Active", nil)
}

func (l *LocationFileLine) SetRemoved() error {
	return l.sendRuleStatus("Deleted", nil)
}

func (l *LocationFileLine) SetError(err rookoutErrors.RookoutError) error {
	return l.sendRuleStatus("Error", err)
}

func NewLocationFileLine(arguments types.AugConfiguration, output com_ws.Output, aug augs.Aug) (Location, rookoutErrors.RookoutError) {
	var location LocationFileLine

	location.filename = arguments["filename"].(string)
	location.lineno = int(arguments["lineno"].(float64))
	location.aug = aug
	location.output = output
	
	return &location, nil
}

func (l *LocationFileLine) GetLineno() int {
	return l.lineno
}

func (l *LocationFileLine) GetFileName() string {
	return l.filename
}

func (l *LocationFileLine) GetAug() augs.Aug {
	return l.aug
}

func (l *LocationFileLine) GetAugId() types.AugId {
	return l.aug.GetAugId()
}
