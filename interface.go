//go:build go1.15 && !go1.20 && cgo
// +build go1.15,!go1.20,cgo

package rookout

import (
	"github.com/Rookout/GoSDK/pkg"
	_ "github.com/Rookout/GoSDK/pkg/services/instrumentation/hooker"
)

type RookOptions = pkg.RookOptions

func Start(opts RookOptions) error {
	start(opts)
	return nil
}

func Stop() {
	stop()
}

func Flush() {
	flush()
}
