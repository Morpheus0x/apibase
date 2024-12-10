package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"gopkg.cc/apibase/config"
	"gopkg.cc/apibase/log"
)

func PostgresInit(pgc config.PostgresConfig, common config.CommonConfig) (config.DB, *log.Error) {
	ssl := ""
	if pgc.SSLMode {
		ssl = "enable"
	} else {
		ssl = "disable"
	}
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", pgc.Host, pgc.Port, pgc.User, pgc.Password, pgc.DB, ssl)
	// connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pgc.User, pgc.Password, pgc.Host, pgc.Port, pgc.DB) // ?ssl=%s , ssl)
	var dbErr error
	for attempt := 1; attempt <= int(common.DB_MAX_RECONNECT_ATTEMPTS); attempt++ {
		var conn *sql.DB
		conn, dbErr = sql.Open("postgres", connString)
		if dbErr != nil {
			log.Logf(log.LevelInfo, "Connecting to database failed, attempt %d/%d", attempt, common.DB_MAX_RECONNECT_ATTEMPTS)
			time.Sleep(common.DB_RECONNECT_TIMEOUT_DURATION())
			continue
		}
		return config.DB{Kind: config.PostgreSQL, Postgres: conn}, log.ErrorNil()
	}
	return config.DB{}, log.NewError(dbErr.Error())
}
