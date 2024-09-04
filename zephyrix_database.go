package zephyrix

func (z *zephyrix) Database() Database {
	return z.db
}

func (z *zephyrix) RegisterEntity(entity ...interface{}) {
	z.db.RegisterEntity(entity...)
}
