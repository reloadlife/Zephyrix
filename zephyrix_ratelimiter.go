package zephyrix

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/latolukasz/beeorm/v3"
	"go.uber.org/fx"
	"golang.org/x/time/rate"
)

// RateLimiterPool represents the configuration for a single rate limiter pool.
type RateLimiterPool struct {
	Name       string        `mapstructure:"name"`
	Limit      rate.Limit    `mapstructure:"limit"`
	Burst      int           `mapstructure:"burst"`
	ExpireTime time.Duration `mapstructure:"expire_time"`
}

// RateLimiterConfig represents the overall configuration for rate limiting.
type RateLimiterConfig struct {
	RedisPool  string            `mapstructure:"redis_pool"`
	LimitPools []RateLimiterPool `mapstructure:"pools"`
}

// RateLimiter manages rate limiting across multiple pools.
type RateLimiter struct {
	config   RateLimiterConfig
	client   beeorm.RedisCache
	orm      beeorm.ORM
	limiters sync.Map
	mu       sync.RWMutex
}

// NewRateLimiter creates a new RateLimiter instance.
//
// It takes a beeorm.Engine and a Config as parameters and returns a pointer to a RateLimiter.
//
// Example:
//
//	engine := beeorm.Engine{} // you have to create an instance of it, obviously.
//	config := &Config{
//		RateLimiter: RateLimiterConfig{
//			RedisPool: "default",
//			LimitPools: []RateLimiterPool{
//				{Name: "default", Limit: 10, Burst: 20, ExpireTime: time.Minute},
//			},
//		},
//	}
//	rateLimiter := NewRateLimiter(engine, config)
func NewRateLimiter(orm beeorm.Engine, config *Config) *RateLimiter {
	conf := config.RateLimiter
	return &RateLimiter{
		config:   conf,
		client:   orm.Redis(conf.RedisPool),
		orm:      orm.NewORM(context.Background()),
		limiters: sync.Map{},
	}
}

// invokeRateLimiter sets up the rate limiter with the fx lifecycle.
//
// It starts a goroutine for cleanup when the application starts.
func invokeRateLimiter(lc fx.Lifecycle, rl *RateLimiter) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go rl.cleanup(ctx)
			return nil
		},
	})
}

// getPoolConfig retrieves the configuration for a specific rate limiter pool.
//
// If the pool name is empty, it returns the configuration for the "default" pool.
// If no matching pool is found, it returns nil.
func (rl *RateLimiter) getPoolConfig(pool string) *RateLimiterPool {
	if pool == "" {
		pool = "default"
	}

	rl.mu.RLock()
	defer rl.mu.RUnlock()

	for _, p := range rl.config.LimitPools {
		if p.Name == pool {
			return &p
		}
	}
	return nil
}

// Limiter creates a new Limiter instance for a specific pool.
//
// If the specified pool doesn't exist, it falls back to the default pool.
//
// Example:
//
//	ctx := context.Background()
//	limiter := rateLimiter.Limiter(ctx, "api")
//
// for this to work, you need the "api" pool in the zephyrix.yaml file.
// Example:
//
//		rate_limiter:
//	  redis_pool: "default"
//	  pools:
//	    - name: "default"
//	      limit: 100
//	      burst: 10
//	      interval: "1m"
//	    - name: "api"
//	      limit: 10
//	      burst: 5
//	      interval: "5m"
func (rl *RateLimiter) Limiter(ctx context.Context, pool string) *Limiter {
	p := rl.getPoolConfig(pool)
	if p == nil {
		p = rl.getPoolConfig("")
	}
	return &Limiter{
		rl:         rl,
		limiter:    rate.NewLimiter(p.Limit, p.Burst),
		Burst:      p.Burst,
		ExpireTime: p.ExpireTime,
	}
}

// cleanup periodically clears the in-memory limiters to prevent memory leaks.
func (rl *RateLimiter) cleanup(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.limiters.Range(func(key, value interface{}) bool {
				rl.limiters.Delete(key)
				return true
			})
		}
	}
}

// Limiter represents a rate limiter for a specific action and key.
type Limiter struct {
	rl         *RateLimiter
	limiter    *rate.Limiter
	Burst      int
	ExpireTime time.Duration
}

// Allow checks if an action is allowed for a given key.
//
// It first checks the local in-memory limiter, and if that allows the action,
// it then checks the Redis-based limiter for distributed rate limiting.
//
// Example:
//
//	if limiter.Allow(ctx, "login", userIP) {
//		// Process login
//	} else {
//		// Return rate limit exceeded error
//	}
func (l *Limiter) Allow(ctx context.Context, action, key string) bool {
	limiterKey := fmt.Sprintf("ratelimiter:%s:%s", action, key)

	if l.allowLocal(limiterKey) {
		return true
	}

	return l.allowRedis(ctx, limiterKey)
}

// allowLocal checks if the action is allowed based on the local in-memory limiter.
func (l *Limiter) allowLocal(key string) bool {
	return l.limiter.Allow()
}

// allowRedis checks if the action is allowed based on the Redis-based limiter.
func (l *Limiter) allowRedis(ctx context.Context, key string) bool {
	l.rl.mu.Lock()
	defer l.rl.mu.Unlock()

	current, _ := l.rl.client.Get(l.rl.orm, key)
	count, _ := strconv.Atoi(current)

	if count >= l.Burst {
		return false
	}

	l.rl.client.Set(l.rl.orm, key, strconv.Itoa(count+1), l.ExpireTime)
	return true
}
