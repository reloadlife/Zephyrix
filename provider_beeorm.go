// Package zephyrix provides a high-performance web framework for Go.
package zephyrix

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/latolukasz/beeorm/v3"
	"github.com/latolukasz/beeorm/v3/plugins/modified"
	"go.uber.org/fx"
)

// DatabasePoolConfig represents the configuration for a single database pool.
type DatabasePoolConfig struct {
	Name              string      `mapstructure:"name"`                // Name of the database pool
	DSN               string      `mapstructure:"dsn"`                 // Data Source Name for the database connection
	MaxOpenConns      int         `mapstructure:"max_open_conns"`      // Maximum number of open connections to the database
	MaxIdleConns      int         `mapstructure:"max_idle_conns"`      // Maximum number of idle connections in the pool
	ConnMaxLifetime   string      `mapstructure:"conn_max_lifetime"`   // Maximum amount of time a connection may be reused
	UnsafeAutoMigrate bool        `mapstructure:"unsafe_auto_migrate"` // Whether to automatically run migrations (use with caution)
	IgnoredTables     []string    `mapstructure:"ignored_tables"`      // Tables to be ignored by the ORM
	DefaultEncoding   string      `mapstructure:"default_encoding"`    // Default character encoding for the database
	DefaultCollate    string      `mapstructure:"default_collation"`   // Default collation for the database
	Cache             CacheConfig `mapstructure:"cache"`               // Configuration for the local cache
	Redis             RedisConfig `mapstructure:"redis"`               // Configuration for Redis cache
}

// CacheConfig represents the configuration for local caching.
type CacheConfig struct {
	Enabled bool `mapstructure:"enabled"` // Whether local caching is enabled
	Size    int  `mapstructure:"size"`    // Size of the local cache
}

// RedisConfig represents the configuration for Redis caching.
type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`  // Whether Redis caching is enabled
	Address  string `mapstructure:"address"`  // Address of the Redis server
	Password string `mapstructure:"password"` // Password for Redis authentication
	Username string `mapstructure:"username"` // Username for Redis authentication
	DB       int    `mapstructure:"db"`       // Redis database number
}

// DatabaseConfig represents the overall database configuration,
// which can include multiple database pools.
type DatabaseConfig struct {
	Pools []DatabasePoolConfig `mapstructure:"pools"` // List of database pool configurations
}

// beeormEngine encapsulates the BeeORM engine and its associated data.
type beeormEngine struct {
	r       beeorm.Registry // BeeORM registry
	e       beeorm.Engine   // BeeORM engine
	models  sync.Map        // Thread-safe map to store registered models
	isDirty bool            // Flag to indicate if the engine needs revalidation
	mu      sync.RWMutex    // Mutex for thread-safe operations
}

// newBeeormEngine creates and initializes a new beeormEngine.
func newBeeormEngine() *beeormEngine {
	r := beeorm.NewRegistry()
	r.RegisterPlugin(modified.New("CreatedAt", "ModifiedAt"))
	return &beeormEngine{
		r:       r,
		isDirty: true,
	}
}

// RegisterEntity registers one or more entities with the BeeORM engine.
// It's safe to call this method concurrently from multiple goroutines.
func (b *beeormEngine) RegisterEntity(entities ...interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, entity := range entities {
		key := fmt.Sprintf("%T", entity)
		b.models.Store(key, entity)
		Logger.Debug("Registered entity: %s", key)
	}
	b.r.RegisterEntity(entities...)
	b.isDirty = true
}

// GetEngine returns the BeeORM engine, validating it if necessary.
// It's safe to call this method concurrently from multiple goroutines.
func (b *beeormEngine) GetEngine() beeorm.Engine {
	b.mu.RLock()
	if !b.isDirty {
		defer b.mu.RUnlock()
		return b.e
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.isDirty {
		return b.e
	}

	engine, err := b.r.Validate()
	if err != nil {
		Logger.Fatal("Failed to validate beeorm engine: %s", err)
		return nil
	}
	b.e = engine
	b.isDirty = false
	return engine
}

// beeormProvider is a factory function that creates a new beeormEngine.
// It's designed to be used with the fx dependency injection framework.
func beeormProvider() *beeormEngine {
	return newBeeormEngine()
}

// beeormInvoke sets up the BeeORM engine with the provided configuration and lifecycle hooks.
// It's designed to be used with the fx dependency injection framework.
func beeormInvoke(lc fx.Lifecycle, bee *beeormEngine, conf *Config) {
	for _, pool := range conf.Database.Pools {
		configurePool(bee.r, pool)
	}

	engine := bee.GetEngine()
	if engine == nil {
		Logger.Fatal("Failed to get beeorm engine")
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go runFlusher(engine)
			return nil
		},
		OnStop: func(context.Context) error {
			// Implement any necessary cleanup here
			return nil
		},
	})
}

// configurePool sets up a single database pool with the given configuration.
func configurePool(r beeorm.Registry, pool DatabasePoolConfig) {
	maxOpenConns := getMaxOpenConns(pool.MaxOpenConns)
	maxIdleConns := getMaxIdleConns(pool.MaxIdleConns)
	maxLifeTime := getMaxLifeTime(pool.ConnMaxLifetime)

	r.RegisterMySQL(pool.DSN, pool.Name, &beeorm.MySQLOptions{
		MaxOpenConnections: maxOpenConns,
		MaxIdleConnections: maxIdleConns,
		ConnMaxLifetime:    maxLifeTime,
		DefaultEncoding:    pool.DefaultEncoding,
		DefaultCollate:     pool.DefaultCollate,
		IgnoredTables:      pool.IgnoredTables,
	})

	if pool.Cache.Enabled {
		r.RegisterLocalCache(pool.Name, pool.Cache.Size)
	}

	if pool.Redis.Enabled {
		r.RegisterRedis(pool.Redis.Address, pool.Redis.DB, pool.Name, &beeorm.RedisOptions{
			Password: pool.Redis.Password,
			User:     pool.Redis.Username,
		})
	}
}

// getMaxOpenConns returns the maximum number of open connections,
// using a default value if none is specified.
func getMaxOpenConns(maxOpenConns int) int {
	if maxOpenConns == 0 {
		return 10
	}
	return maxOpenConns
}

// getMaxIdleConns returns the maximum number of idle connections,
// using a default value if none is specified.
func getMaxIdleConns(maxIdleConns int) int {
	if maxIdleConns == 0 {
		return 5
	}
	return maxIdleConns
}

// getMaxLifeTime parses and returns the maximum connection lifetime,
// using a default value if none is specified or if parsing fails.
func getMaxLifeTime(connMaxLifetime string) time.Duration {
	if connMaxLifetime == "" {
		return 30 * time.Minute
	}
	maxLifeTime, err := time.ParseDuration(connMaxLifetime)
	if err != nil {
		Logger.Warn("Invalid ConnMaxLifetime: %s. Using default of 30 minutes.", connMaxLifetime)
		return 30 * time.Minute
	}
	return maxLifeTime
}

// runFlusher periodically flushes asynchronous database events.
// It runs in its own goroutine.
func runFlusher(engine beeorm.Engine) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c := engine.NewORM(context.Background())
		c.SetMetaData("source", "flushed_by_zephyrix_consumer")
		Logger.Debug("Flusher is running")
		if err := beeorm.ConsumeAsyncFlushEvents(c, true); err != nil {
			Logger.Error("Failed to flush database events: %s", err)
		} else {
			Logger.Debug("Database events flushed")
		}
	}
}
