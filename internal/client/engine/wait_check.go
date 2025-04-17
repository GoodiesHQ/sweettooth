package engine

import (
	"time"

	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/api/client"
)

const (
	DEFAULT_PERIOD_WAITCHECK = time.Second * 10 // time to wait while waiting for admin approval
)

func (engine *SweetToothEngine) WaitCheck() bool {
	log := util.Logger("engine.WaitCheck")
	log.Trace().Msg("called")
	defer log.Trace().Msg("finish")

	for engine.isRunning() {
		if err := engine.client.Check(); err == nil {
			return true
		} else if err == client.ErrNodeNotRegistered {
			engine.client.Registered = false
			log.Panic().Msg("node is no longer registered")
		}

		select {
		case <-time.After(DEFAULT_PERIOD_WAITCHECK):
			continue
		case <-engine.GetStopChan():
			panic(ErrStop)
		}
	}

	// engine is no longer running
	panic(ErrStop)
}
