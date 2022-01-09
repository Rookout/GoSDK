package rookout

import (
	"github.com/Rookout/GoSDK/pkg"
)

type RookOptions = pkg.RookOptions

func Start(opts RookOptions) error {
	return start(opts)
}

func Stop() {
	stop()
}
