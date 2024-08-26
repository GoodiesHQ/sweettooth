package engine

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/rs/zerolog/log"
)

var ErrStop = errors.New("engine has been stopped")

type SweetToothEngine struct {
	client       *client.SweetToothClient
	bootstrapped bool
	running      atomic.Bool
	mu           sync.Mutex
	stopch       chan bool
	wg           sync.WaitGroup
}

func NewSweetToothEngine(client *client.SweetToothClient) *SweetToothEngine {
	return &SweetToothEngine{
		client: client,
		mu:     sync.Mutex{},
		stopch: nil,
		wg:     sync.WaitGroup{},
	}
}

func (engine *SweetToothEngine) Wait() {
	log := log.Logger.With().Str("routine", "enigne.Wait").Logger()
	log.Trace().Msg("called")
	defer log.Trace().Msg("finished")

	if ch := engine.GetStopChan(); ch != nil {
		log.Trace().Msg("waiting...")
		<-ch
		log.Trace().Msg("done waiting")
	}
}

func (engine *SweetToothEngine) Start() {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	log := log.Logger.With().Str("routine", "enigne.Start").Logger()
	log.Trace().Msg("called")
	defer log.Trace().Msg("finished")

	if engine.isRunning() {
		log.Trace().Msg("engine is already running")
		return
	}

	engine.stopch = make(chan bool, 1)
	engine.wg.Add(1)

	go func() {
		log.Trace().Msg("goroutine executing engine.Run()")
		defer engine.wg.Done()
		engine.run()
		log.Trace().Msg("goroutine finished engine.Run()")
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
