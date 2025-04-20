package cache

import (
	"time"
)

// Cache interface used by sweettooth
type Cache interface {
	SetAuthWithLifetime(nodeid string, authorized bool, lifetime time.Duration)
	SetNodeAuth(nodeid string, authorized bool)                  // set node authorization status
	GetNodeAuth(nodeid string) (found bool, authorized bool) // get node authorization status
	Flush()                                                  // clear the cache
}

func CacheSuffixAuth(s string) string {
	return s + "-auth"
}
