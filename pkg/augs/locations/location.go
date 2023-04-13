package locations

import (
	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/types"
)

type Location interface {
	GetLineno() int
	GetFileName() string
	GetAug() augs.Aug
	GetAugID() types.AugID
	SetPending() error
	SetActive() error
	SetRemoved() error
	SetError(rookoutErrors.RookoutError) error
}
