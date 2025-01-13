package db

import (
	"time"

	h "gopkg.cc/apibase/helper"
)

type BaseConfig struct {
	DB_RECONNECT_TIMEOUT_SEC  uint   `toml:"db_reconnect_timeout_sec"`
	DB_MAX_RECONNECT_ATTEMPTS uint   `toml:"db_max_reconnect_attempts"`
	SQLITE_DATETIME_FORMAT    string `toml:"sqlite_datetime_format"`
}

func (bc *BaseConfig) DB_RECONNECT_TIMEOUT_DURATION() time.Duration {
	return time.Second * time.Duration(bc.DB_RECONNECT_TIMEOUT_SEC)
}

func GetBaseConfigDefault() BaseConfig {
	return BaseConfig{
		DB_RECONNECT_TIMEOUT_SEC:  1,
		DB_MAX_RECONNECT_ATTEMPTS: 3,
	}
}

type PostgresConfig struct {
	Host     string         `toml:"host"`
	Port     string         `toml:"port"`
	User     string         `toml:"user"`
	Password h.SecretString `toml:"password"`
	DB       string         `toml:"db"`
	SSLMode  bool           `toml:"ssl_enabled"`
}

type SQLiteConfig struct {
	FilePath string `toml:"file_path"`
	LockFile string `toml:"lock_file"`
}
