package zephyrix

import (
	"github.com/latolukasz/beeorm/v3"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	runUnsafeMigrations bool = false
)

func (z *zephyrix) migrationRun(_ *cobra.Command, _ []string) error {
	z.options = append(z.options, fx.Invoke(z.migrate))
	err := z.fxStart()
	if err != nil {
		return err
	}
	return nil
}

type Database interface {
	RegisterEntity(entity ...interface{})
	GetEngine() beeorm.Engine
}

func (z *zephyrix) Database() Database {
	return z.db
}

func (z *zephyrix) migrate() {
	c := z.db.GetEngine().NewORM(z.c)
	alters := beeorm.GetAlters(c)
	for _, alter := range alters {
		if !alter.Safe {
			Logger.Warn("Unsafe migration detected: %s (pool: %s)", alter.SQL, alter.Pool)
			if !runUnsafeMigrations {
				Logger.Warn("To run unsafe migrations, use the --unsafe-migrations flag")
				continue
			}
		}
		alter.Exec(c)
	}

}
