package runtime

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type securityCooldown struct {
	cache    *cache.Cache
	cooldown time.Duration
}

func newSecurityCooldown(cooldown time.Duration) *securityCooldown {
	return &securityCooldown{
		cache: cache.New(cooldown, time.Duration(int64(1.3*float64(cooldown)))),
	}
}

func (c *securityCooldown) CanRunNow(runtimeId string, enabled bool) bool {
	val, found := c.cache.Get(runtimeId)
	if !found {
		// not called recently, enough time has passed, ok to run now
		c.cache.Set(runtimeId, enabled, cache.DefaultExpiration)
		return true
	}
	cachedEnabled := val.(bool)
	if cachedEnabled != enabled {
		// was run in recent pass (less than cooldown) but with different enabled state
		// ok to run now since state is different
		c.cache.Set(runtimeId, enabled, cache.DefaultExpiration)
		return true
	}
	// was run in recent pass with the same enabled state, not ok to run now
	return false
}

func (c *securityCooldown) MarkHasRun(runtimeId string, enabled bool) {
	c.cache.Set(runtimeId, enabled, cache.DefaultExpiration)
}
