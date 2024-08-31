package zephyrix

import "github.com/spf13/cobra"

func (z *zephyrix) migrationRun(_ *cobra.Command, _ []string) error {
	// z.options = append(z.options, fx.Invoke(migrate))
	err := z.fxStart()
	if err != nil {
		return err
	}
	return nil
}

type Database interface {
	RegisterEntity(entity ...interface{})
}

func (z *zephyrix) Database() Database {
	return z.db
}
