package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	REDIS_TRUE  = "1"
	REDIS_FALSE = "0"
)

type CacheRedis struct {
	c        *redis.Client
	lifetime time.Duration
}

func NewCacheRedis(host string, port uint16, password string, lifetime time.Duration) *CacheRedis {
	return &CacheRedis{
		c: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: password,
		}),
		lifetime: lifetime,
	}
}

// Set the authorization status of a node (cache-only) using an explicit expiration time
func (c *CacheRedis) SetAuthWithLifetime(nodeid string, authorized bool, lifetime time.Duration) {
	var isAuthorized string

	if authorized {
		isAuthorized = REDIS_TRUE
	} else {
		isAuthorized = REDIS_FALSE
	}

	c.c.Set(context.Background(), CacheSuffixAuth(nodeid), isAuthorized, lifetime)
}

// Set the authorization status of a node (cache-only) using the default expiration time
func (c *CacheRedis) SetAuth(nodeid string, authorized bool) {
	c.SetAuthWithLifetime(nodeid, authorized, c.lifetime)
}

// Get the auth status of a Node ID (only reliable if `found` is true, meaning it was found in the cache and the value can be trusted)
func (c *CacheRedis) GetAuth(nodeid string) (found, authorized bool) {
	isAuthorized, err := c.c.Get(context.Background(), CacheSuffixAuth(nodeid)).Result()
	if err == redis.Nil {
		return false, false
	} else if err != nil {
		log.Warn().Err(err).Msg("redis cache failure")
		return false, false
	}

	return true, isAuthorized == REDIS_TRUE
}

// Flush the cache
func (c *CacheRedis) Flush() {
	c.c.FlushDB(context.Background())
}
