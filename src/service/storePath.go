package service

import (
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/model"
	"gorm.io/gorm"
)

type StorePathSrv struct {
	db *gorm.DB
}

func NewStorePathSrv() *StorePathSrv {
	var srv *StorePathSrv

	app.Provide(func(db *gorm.DB) {
		srv = &StorePathSrv{db}
	})

	return srv
}

func (s *StorePathSrv) ReadAll(cache string) ([]model.StorePath, error) {
	// TODO: implement
	panic("unimplemented")
}

func (s *StorePathSrv) Read(storeHash string, cache string) (*model.StorePath, error) {
	// TODO: implement
	panic("unimplemented")
}

func (s *StorePathSrv) Delete(storeHash string, cache string) error {
	// TODO: implement
	panic("unimplemented")
}
