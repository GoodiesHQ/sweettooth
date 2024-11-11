package engine

import (
	"time"

	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/rs/zerolog/log"
)

const (
	DEFAULT_PERIOD_WAITCHECK = time.Second * 10 // time to wait while waiting for admin approval
)

func (engine *SweetToothEngine) WaitCheck() bool {
	log.Trace().Str("routine", "WaitCheck").Msg("called")
	defer log.Trace().Str("routine", "WaitCheck").Msg("finished")

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
