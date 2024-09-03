package zephyrix

import (
	"context"
	"fmt"
	"time"

	"github.com/latolukasz/beeorm/v3"
	"github.com/latolukasz/beeorm/v3/plugins/modified"
	"go.uber.org/fx"
)

type DatabasePoolConfig struct {
	Name              string      `mapstructure:"name"`
	DSN               string      `mapstructure:"dsn"`
	MaxOpenConns      int         `mapstructure:"max_open_conns"`
	MaxIdleConns      int         `mapstructure:"max_idle_conns"`
	ConnMaxLifetime   string      `mapstructure:"conn_max_lifetime"`
	UnsafeAutoMigrate bool        `mapstructure:"unsafe_auto_migrate"`
	IgnoredTables     []string    `mapstructure:"ignored_tables"`
	DefaultEncoding   string      `mapstructure:"default_encoding"`
	DefaultCollate    string      `mapstructure:"default_collation"`
	Cache             CacheConfig `mapstructure:"cache"`
	Redis             RedisConfig `mapstructure:"redis"`
}

type CacheConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Size    int  `mapstructure:"size"`
}

type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	Username string `mapstructure:"username"`
	DB       int    `mapstructure:"db"`
}

type DatabaseConfig struct {
	Pools []DatabasePoolConfig `mapstructure:"pools"`
}

type beeormEngine struct {
	r beeorm.Registry
	e beeorm.Engine

	models  map[string]interface{}
	isDirty bool
}

func beeormProvider() *beeormEngine {
	bee := beeormEngine{
		models: make(map[string]interface{}),
	}
	r := beeorm.NewRegistry()
	r.RegisterPlugin(modified.New("CreatedAt", "ModifiedAt"))
	bee.r = r
	bee.isDirty = true
	return &bee
}

func (b *beeormEngine) RegisterEntity(entity ...interface{}) {
	Logger.Debug("RegisterEntity %s %#v", "entity", entity)
	for _, e := range entity {
		b.models[fmt.Sprintf("%T", e)] = e
	}
	b.r.RegisterEntity(entity...)
	b.isDirty = true
}

func (b *beeormEngine) GetEngine() beeorm.Engine {
	if !b.isDirty {
		return b.e
	}
	r := b.r
	engine, err := r.Validate()
	if err != nil {
		Logger.Fatal("Failed to validate beeorm engine: %s", err)
		return nil
	}
	b.e = engine
	return engine
}

func beeormInvoke(lc fx.Lifecycle, bee *beeormEngine, conf *Config) {
	r := bee.r

	config := conf.Database
	for _, pool := range config.Pools {
		if pool.MaxOpenConns == 0 {
			pool.MaxOpenConns = 10
		}
		if pool.MaxIdleConns == 0 {
			pool.MaxIdleConns = 5
		}
		if pool.ConnMaxLifetime == "" {
			pool.ConnMaxLifetime = "30m"
		}

		maxLifeTime, _ := time.ParseDuration(pool.ConnMaxLifetime)
		r.RegisterMySQL(pool.DSN, pool.Name, &beeorm.MySQLOptions{
			MaxOpenConnections: pool.MaxOpenConns,
			MaxIdleConnections: pool.MaxIdleConns,
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

	engine := bee.GetEngine()
	if engine == nil {
		Logger.Fatal("Failed to get beeorm engine")
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				c := engine.NewORM(context.Background())
				c.SetMetaData("source", "flushed_by_zephyrix_consumer")
				for {
					Logger.Debug("Flusher is running")
					flushError := beeorm.ConsumeAsyncFlushEvents(c, true)
					if flushError != nil {
						Logger.Error("Failed to flush database Events! %s", flushError)
					}
					Logger.Debug("Database events Flushed")
					time.Sleep(time.Minute)
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			return nil
		},
	})
}
