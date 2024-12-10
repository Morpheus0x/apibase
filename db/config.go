package db

type PostgresConfig struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DB       string `toml:"db"`
	SSLMode  bool   `toml:"ssl_enabled"`
}

type SQLiteConfig struct {
	FilePath string `toml:"file_path"`
	LockFile string `toml:"lock_file"`
}
