package rookout

import (
	"github.com/Rookout/Go/pkg"
)

type RookOptions = pkg.RookOptions

func Start(opts RookOptions) error {
	return start(opts)
}

func Stop() {
	stop()
}
