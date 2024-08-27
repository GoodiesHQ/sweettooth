package engine

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/rs/zerolog/log"
)

var ErrStop = errors.New("engine has been stopped")

type SweetToothEngine struct {
	client       *client.SweetToothClient
	config       *config.Configuration
	bootstrapped bool
	running      atomic.Bool
	mu           sync.Mutex
	stopch       chan bool
	wg           sync.WaitGroup
}

// create a new engine from a configuration file
func NewSweetToothEngine(cfg *config.Configuration) *SweetToothEngine {
	return &SweetToothEngine{
		client:  client.NewSweetToothClient(cfg.Server.Url),
		config:  cfg,
		mu:      sync.Mutex{},
		stopch:  nil,
		running: atomic.Bool{},
		wg:      sync.WaitGroup{},
	}
}

// wait for all goroutines to complete (for stop to be called)
func (engine *SweetToothEngine) Wait() {
	log := log.Logger.With().Str("routine", "enigne.Wait").Logger()
	if ch := engine.GetStopChan(); ch != nil {
		log.Trace().Msg("engine is waiting...")
		<-ch
		log.Trace().Msg("engine is done waiting")
	}
}

func (engine *SweetToothEngine) Start() {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	log := log.With().Str("routine", "enigne.Start").Logger()
	log.Trace().Msg("called")
	defer log.Trace().Msg("finished")

	// check if the engine is running, don't bother if it is
	if engine.isRunning() {
		log.Trace().Msg("engine is already running")
		return
	}

	// create a new signal channel and add a worker
	engine.stopch = make(chan bool, 1)
	engine.wg.Add(1)

	// launch the goroutine which performs the actual work
	go func() {
		log := log.With().Str("subroutine", "enigne.Start()::run").Logger()
		log.Trace().Msg("goroutine executing")
		defer engine.wg.Done()
		engine.run()
		log.Trace().Msg("goroutine finished")
	}()

	// officially running
	engine.running.Store(true)
}

// return a read-only stop channel and a cleanup function
func (engine *SweetToothEngine) GetStopChan() <-chan bool {
	return engine.stopch
}

func (engine *SweetToothEngine) isRunning() bool {
	return engine.running.Load()
}

func (engine *SweetToothEngine) Stop() {
	log.Warn().Msg("engine.stop called")
	engine.mu.Lock()
	defer engine.mu.Unlock()

	if !engine.running.CompareAndSwap(true, false) {
		log.Warn().Msg("engine was already stopped")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			engine.running.Store(true)
		}
	}()

	log.Warn().Msg("closing internal signal channels")
	if engine.stopch != nil {
		select {
		case <-engine.stopch:
			log.Warn().Msg("engine stopch is already closed")
		default:
			close(engine.stopch)
			engine.running.Store(false)
			engine.Wait()
		}
	}
}
