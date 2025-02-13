package baseconfig

import (
	"time"

	h "gopkg.cc/apibase/helper"
)

type BaseConfig struct {
	DatabaseMaxReconnectAttempts  uint   `toml:"db_max_reconnect_attempts"`
	SQLiteDatetimeFormat          string `toml:"sqlite_datetime_format"`
	TomlTimeoutCloseChainShutdown string `toml:"timeout_closechain_shutdown"`
	TomlTimeoutDatabaseConnect    string `toml:"timeout_database_connect"`
	TomlTimeoutDatabaseShutdown   string `toml:"timeout_database_shutdown"`
	TomlTimeoutDatabaseQuery      string `toml:"timeout_database_query"`
	TomlTimeoutDatabaseLargeQuery string `toml:"timeout_database_large_query"`

	TimeoutCloseChainShutdown time.Duration `internal:"timeout_closechain_shutdown"`
	TimeoutDatabaseConnect    time.Duration `internal:"timeout_database_connect"`
	TimeoutDatabaseShutdown   time.Duration `internal:"timeout_database_shutdown"`
	TimeoutDatabaseQuery      time.Duration `internal:"timeout_database_query"`
	TimeoutDatabaseLargeQuery time.Duration `internal:"timeout_database_large_query"`
}

func (bc *BaseConfig) AddMissingFromDefaults() error {
	defaults := &BaseConfig{
		DatabaseMaxReconnectAttempts: 3,
		SQLiteDatetimeFormat:         "2006-01-02 15:04:05",

		TimeoutCloseChainShutdown: time.Second * 30,
		TimeoutDatabaseConnect:    time.Second * 3,
		TimeoutDatabaseShutdown:   time.Second * 3,
		TimeoutDatabaseQuery:      time.Second * 10,
		TimeoutDatabaseLargeQuery: time.Minute * 5,
	}
	return h.ParseTomlConfigAndDefaults(bc, defaults)
}
