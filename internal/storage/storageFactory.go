package storage

import (
	"github.com/boltdb/bolt"
	"github.com/t-bfame/diago/internal/daoImplBoltDB"
)

type Factory struct {
	db           *bolt.DB
	isPersistent bool
}

func NewStorageFactory(filePath string, isPersistent bool) (*Factory, error) {
	if !isPersistent {
		return nil, nil //todo
	}

	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		return nil, err
	}
	newStorageFactory := Factory{db: db, isPersistent: isPersistent}
	return &newStorageFactory, nil
}

func (sf *Factory) GetJobDao() JobDao {
	if sf.isPersistent {
		result := daoImplBoltDB.NewJobDao(sf.db)
		return result
	} else {
		return nil //todo
	}
}

func (sf *Factory) GetTestDao() TestDao {
	if sf.isPersistent {
		result := daoImplBoltDB.NewTestDao(sf.db)
		return result
	} else {
		return nil //todo
	}
}

func (sf *Factory) GetTestInstanceDao() TestInstanceDao {
	if sf.isPersistent {
		result := daoImplBoltDB.NewTestInstanceDao(sf.db)
		return result
	} else {
		return nil //todo
	}
}
