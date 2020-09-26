package storage

var (
	daoFactory *Factory
)

func InitDatabase(isPersistent bool) {
	factory, err := NewStorageFactory("my.db", isPersistent)
	daoFactory = factory
	if err != nil {
		panic(err)
	}
}
