package util

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type RWMutexLogged struct {
	mu sync.RWMutex
}

func (mu *RWMutexLogged) Lock(name string) {
	log.Trace().Str("name", name).Str("type", "write").Msg("locking")
	mu.mu.Lock()
	log.Trace().Str("name", name).Str("type", "write").Msg("locked")
}

func (mu *RWMutexLogged) Unlock(name string) {
	log.Trace().Str("name", name).Str("type", "write").Msg("unlocking")
	mu.mu.Unlock()
	log.Trace().Str("name", name).Str("type", "write").Msg("unlocked")
}

func (mu *RWMutexLogged) RLock(name string) {
	log.Trace().Str("name", name).Str("type", "read").Msg("locking")
	mu.mu.Lock()
	log.Trace().Str("name", name).Str("type", "read").Msg("locked")
}

func (mu *RWMutexLogged) RUnlock(name string) {
	log.Trace().Str("name", name).Str("type", "read").Msg("unlocking")
	mu.mu.Unlock()
	log.Trace().Str("name", name).Str("type", "read").Msg("unlocked")
}
