package locations

import (
	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/com_ws"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
)

type Location interface {
	GetStatus() string
	GetOutput() com_ws.Output
	GetLineno() int
	GetFileName() string
	GetAug() augs.Aug
	GetAugId() types.AugId
	SetPending() error
	SetActive() error
	SetRemoved() error
	SetError(rookoutErrors.RookoutError) error
}
