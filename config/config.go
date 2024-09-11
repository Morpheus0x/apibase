package config

import "time"

type BaseConfig struct {
	DB_RECONNECT_TIMEOUT_SEC  uint `toml:"db_reconnect_timeout_sec"`
	DB_MAX_RECONNECT_ATTEMPTS uint `toml:"db_max_reconnect_attempts"`
}

func (bc *BaseConfig) DB_RECONNECT_TIMEOUT_DURATION() time.Duration {
	return time.Second * time.Duration(bc.DB_RECONNECT_TIMEOUT_SEC)
}

func GetDefaultBaseConfig() BaseConfig {
	return BaseConfig{
		DB_RECONNECT_TIMEOUT_SEC:  1,
		DB_MAX_RECONNECT_ATTEMPTS: 3,
	}
}
