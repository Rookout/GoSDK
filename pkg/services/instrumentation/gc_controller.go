package instrumentation

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/Rookout/GoSDK/pkg/logger"
)

type gcController struct {
	gcEnabled         bool
	triggerCounter    int
	enableGCTimer     <-chan time.Time
	originalGCPercent int
}

func newGCController() *gcController {
	originalGCPercent := debug.SetGCPercent(-1)
	debug.SetGCPercent(originalGCPercent)

	return &gcController{
		gcEnabled:         true,
		triggerCounter:    0,
		enableGCTimer:     nil,
		originalGCPercent: originalGCPercent,
	}
}

func (g *gcController) start(ctx context.Context, triggerChan chan bool) {
	for {
		select {
		case <-ctx.Done():
			g.enableGC()
			return
		case triggerStart := <-triggerChan:
			if triggerStart {
				if g.triggerStart() {
					g.disableGC()
				}
			} else {
				if g.triggerEnd() {
					g.startEnableGCTimer()
				}
			}
		
		case <-g.enableGCTimer:
			g.enableGC()
		}
	}
}

func (g *gcController) triggerStart() bool {
	g.triggerCounter++
	
	return true
}

func (g *gcController) triggerEnd() bool {
	g.triggerCounter--

	
	if g.triggerCounter < 0 {
		logger.Logger().Fatalf("Something bad happened. Trigger counter is a negative number: %d\n", g.triggerCounter)
		g.triggerCounter = 0
		return true
	}

	
	return g.triggerCounter == 0
}

func (g *gcController) disableGC() {
	prevGCPercent := debug.SetGCPercent(-1)
	if prevGCPercent != -1 && prevGCPercent != g.originalGCPercent {
		logger.Logger().Infof("GC percent was changed from %d to %d", g.originalGCPercent, prevGCPercent)
		g.originalGCPercent = prevGCPercent
	}
	g.enableGCTimer = nil
}

func (g *gcController) enableGC() {
	debug.SetGCPercent(g.originalGCPercent)
}

func (g *gcController) startEnableGCTimer() {
	g.enableGCTimer = time.After(1 * time.Millisecond)
}
