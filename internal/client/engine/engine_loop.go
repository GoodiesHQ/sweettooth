package engine

import (
	"errors"
	"fmt"
	"time"

	"github.com/goodieshq/sweettooth/internal/client/schedule"
	"github.com/goodieshq/sweettooth/internal/client/tracker"
	"github.com/rs/zerolog/log"
)

// asserts that the engine is running, panics with ErrStop if not
func (engine *SweetToothEngine) mustRun() {
	if !engine.isRunning() {
		panic(ErrStop)
	}
}

func (engine *SweetToothEngine) loopOnce() bool {
	// bootstrap the engine and system if it hasn't already, otherwise silently continue
	engine.Bootstrap()

	// register the client if it is not already registered, otherwise silently continue
	// engine.Register(context.Background(), uuid.MustParse(""))

	// wait for the first successful check in (wait for an admin to approve the public key if necessary)
	engine.WaitCheck()
	log.Debug().Msg("successfully checked in")

	// TODO: update sources

	// acquire the schedule for this node, just in case it has changed.
	engine.Schedule()
	log.Debug().Msg("successfully loaded the schedule")

	// check package jobs if it is currently in a maintenance schedule
	if BYPASS_SCHEDULE || schedule.Matches() {
		engine.PackageJobs()
		log.Debug().Msg("successfully performed all package jobs")
	}

	// inventory local software and compare with server's inventory, update if needed
	engine.Tracker()
	log.Debug().Msg("successfully completed the software tracker")

	return true
}

// performs one interation of the logic loop, returns any recovered panic errors
func (engine *SweetToothEngine) loopOnceRecoverable() (err error) {
	// defer util.Recoverable(false)
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		if e, ok := r.(error); ok {
			err = e
			return
		}

		if s, ok := r.(string); ok {
			err = errors.New(s)
			return
		}

		err = fmt.Errorf("UNKNOWNERROR %+v", r)
	}()

	// recover from panics during this iteration of the logic loop
	engine.loopOnce()

	return
}

// main logic loop for the client engine, this function should not panic ever
func (engine *SweetToothEngine) loop() {
	log.Trace().Msg("entering the engine loop")

	var i uint64 = 0
	stopch := engine.GetStopChan()
	for engine.isRunning() {
		i += 1
		log.Trace().Uint64("loops", i).Send()

		if err := engine.loopOnceRecoverable(); err != nil {
			// a panic is by definition unexpected behavior, handle it here

			if err == ErrStop {
				log.Warn().Msg("engine has received a stop signal panic, the loop will not continue.")
				return
			} else {
				log.Error().Err(err).Msg("recovered from an error. Re-starting application loop.")
			}

			engine.client.Registered = false // reset the registration status just in case
			tracker.Reset()                  // reset the software inventory to require a fresh server response

			select {
			case <-time.After(DEFAULT_PERIOD_RECOVER): // sleep for a short time before re-starting
				continue
			case <-engine.stopch:
				log.Warn().Msg("engine was stopped while recovering from an error")
				return
			}
		}
		select {
		case <-time.After(DEFAULT_PERIOD_LOOP):
			continue
		case <-stopch:
			log.Warn().Msg("engine was stopped while waiting for the next loop")
			return
		}
	}
	log.Info().Msg("client has been stopped, exiting the client logic loop")
}
