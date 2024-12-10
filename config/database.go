package config

import (
	"database/sql"
	"time"

	"gopkg.cc/apibase/sqlite"
)

type DBKind uint

const (
	SQLite DBKind = iota
	PostgreSQL
)

type DB struct {
	Kind     DBKind
	SQLite   *sqlite.SQLite
	Postgres *sql.DB
}

type CommonConfig struct {
	DB_RECONNECT_TIMEOUT_SEC  uint   `toml:"db_reconnect_timeout_sec"`
	DB_MAX_RECONNECT_ATTEMPTS uint   `toml:"db_max_reconnect_attempts"`
	SQLITE_DATETIME_FORMAT    string `toml:"sqlite_datetime_format"`
}

type PostgresConfig struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DB       string `toml:"db"`
	SSLMode  bool   `toml:"ssl_enabled"`
}

type SQLiteConfig struct {
	FilePath string
	LockFile string
}

func (bc *CommonConfig) DB_RECONNECT_TIMEOUT_DURATION() time.Duration {
	return time.Second * time.Duration(bc.DB_RECONNECT_TIMEOUT_SEC)
}

func GetCommonConfigDefault() CommonConfig {
	return CommonConfig{
		DB_RECONNECT_TIMEOUT_SEC:  1,
		DB_MAX_RECONNECT_ATTEMPTS: 3,
		SQLITE_DATETIME_FORMAT:    "2006-01-02 15:04:05",
	}
}
