package config

type ApiBase struct {
	CommonConfig CommonConfig
	Postgres     PostgresConfig
	SQLite       SQLiteConfig
	ApiBase      ApiConfig
}
