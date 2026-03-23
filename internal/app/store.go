package app

import (
	"github.com/nunoOliveiraqwe/micro-proxy/internal/sqlite"
	"github.com/nunoOliveiraqwe/micro-proxy/internal/store"
)

type DataStore struct {
	UserStore         store.UserStore
	RoleStore         store.RoleStore
	SystemConfigStore store.SystemConfigStore
	ProxyMetricsStore store.ProxyMetricsStore
}

func NewDataStore(db *sqlite.DB) *DataStore {
	return &DataStore{
		UserStore:         sqlite.NewUserStore(db),
		RoleStore:         sqlite.NewRoleStore(db),
		SystemConfigStore: sqlite.NewSystemConfigStore(db),
		ProxyMetricsStore: sqlite.NewProxyMetricsStore(db),
	}
}
