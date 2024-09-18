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

// DatabaseConfig represents the overall database configuration,
// which can include multiple database pools.
type DatabaseConfig struct {
	Pools []DatabasePoolConfig `mapstructure:"pools"` // List of database pool configurations
}

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
type beeormEngine struct {
	conf   *DatabaseConfig
	r      beeorm.Registry
	e      beeorm.Engine
	models sync.Map
	mu     sync.RWMutex
}

func newBeeormEngine() *beeormEngine {
	r := beeorm.NewRegistry()
	r.RegisterPlugin(modified.New("CreatedAt", "ModifiedAt"))
	return &beeormEngine{
		r: r,
	}
}

// RegisterEntity now only adds entities to the models map
func (b *beeormEngine) RegisterEntity(entities ...interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, entity := range entities {
		key := fmt.Sprintf("%T", entity)
		b.models.Store(key, entity)
		Logger.Debug("Added entity to registration queue: %s", key)
	}
}

func (b *beeormEngine) GetEngine() beeorm.Engine {
	b.mu.RLock()
	if b.e != nil {
		defer b.mu.RUnlock()
		return b.e
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.e != nil {
		return b.e
	}

	engine, err := b.r.Validate()
	if err != nil {
		Logger.Fatal("Failed to validate beeorm engine: %s", err)
		return nil
	}
	b.e = engine
	fmt.Print("BeeORM engine created\n")
	return engine
}

func beeormProvider() *beeormEngine {
	return newBeeormEngine()
}

func beeormInvoke(lc fx.Lifecycle, bee *beeormEngine, conf *Config) {
	bee.conf = &conf.Database

	// Register all entities from the models map
	bee.models.Range(func(_, value interface{}) bool {
		bee.r.RegisterEntity(value)
		return true
	})

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
			return nil // Add any necessary cleanup here
		},
	})
}

func configurePool(r beeorm.Registry, pool DatabasePoolConfig) {
	maxOpenConns := defaultIfZero(pool.MaxOpenConns, 10)
	maxIdleConns := defaultIfZero(pool.MaxIdleConns, 5)
	maxLifeTime := parseMaxLifeTime(pool.ConnMaxLifetime)

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

func defaultIfZero(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

func parseMaxLifeTime(connMaxLifetime string) time.Duration {
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

func (b *beeormEngine) HasPool(name string) bool {
	if b.conf == nil {
		return true // assume it has the pool
	}
	for _, pool := range b.conf.Pools {
		if pool.Name == name {
			return true
		}
	}
	return false
}
