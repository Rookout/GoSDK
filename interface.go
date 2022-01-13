package rookout

import (
	"github.com/Rookout/GoSDKDum/pkg"
)

type RookOptions = pkg.RookOptions

func Start(opts RookOptions) error {
	return start(opts)
}

func Stop() {
	stop()
}
