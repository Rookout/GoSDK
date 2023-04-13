//go:build go1.15 && !go1.21 && cgo && (amd64 || arm64)
// +build go1.15
// +build !go1.21
// +build cgo
// +build amd64 arm64

package rookout

import (
	"github.com/Rookout/GoSDK/pkg/config"
	_ "github.com/Rookout/GoSDK/pkg/services/instrumentation/hooker"
)

type RookOptions = config.RookOptions

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
