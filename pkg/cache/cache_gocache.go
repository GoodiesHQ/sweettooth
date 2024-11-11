package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type CacheGo struct {
	c *cache.Cache
}

func NewCacheGo(lifetime, cleanup time.Duration) *CacheGo {
	return &CacheGo{
		c: cache.New(lifetime, cleanup),
	}
}

// Set the authorization status of a node (cache-only) using an explicit expiration time
func (c *CacheGo) SetAuthWithLifetime(nodeid string, authorized bool, lifetime time.Duration) {
	c.c.Set(CacheSuffixAuth(nodeid), authorized, lifetime)
}

// Set the authorization status of a node (cache-only) using the default expiration time
func (c *CacheGo) SetAuth(nodeid string, authorized bool) {
	c.SetAuthWithLifetime(nodeid, authorized, 0)
}

// Get the auth status of a Node ID (only reliable if `found` is true, meaning it was found in the cache and the value can be trusted)
func (c *CacheGo) GetAuth(nodeid string) (found, authorized bool) {
	isAuthorized, found := c.c.Get(CacheSuffixAuth(nodeid))
	if found {
		authorized = isAuthorized.(bool)
	}
	return
}

// Flush the cache
func (c *CacheGo) Flush() {
	c.c.Flush()
}
