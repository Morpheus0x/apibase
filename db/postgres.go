package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"gopkg.cc/apibase/log"
)

func PostgresInit(pgc PostgresConfig, bc BaseConfig) (DB, *log.Error) {
	ssl := ""
	if pgc.SSLMode {
		ssl = "enable"
	} else {
		ssl = "disable"
	}
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", pgc.Host, pgc.Port, pgc.User, pgc.Password, pgc.DB, ssl)
	// connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pgc.User, pgc.Password, pgc.Host, pgc.Port, pgc.DB) // ?ssl=%s , ssl)
	var dbErr error
	for attempt := 1; attempt <= int(bc.DB_MAX_RECONNECT_ATTEMPTS); attempt++ {
		var conn *sql.DB
		conn, dbErr = sql.Open("postgres", connString)
		if dbErr != nil {
			log.Logf(log.LevelInfo, "Connecting to database failed, attempt %d/%d", attempt, bc.DB_MAX_RECONNECT_ATTEMPTS)
			time.Sleep(bc.DB_RECONNECT_TIMEOUT_DURATION())
			continue
		}
		return DB{Kind: PostgreSQL, Postgres: conn}, log.ErrorNil()
	}
	return DB{}, log.NewError(dbErr.Error())
}
