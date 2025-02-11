package db

import (
	h "gopkg.cc/apibase/helper"
)

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
