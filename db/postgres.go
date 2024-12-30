package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"

	"gopkg.cc/apibase/log"
)

func PostgresInit(pgc PostgresConfig, bc BaseConfig) (DB, *log.Error) {
	db := DB{Kind: PostgreSQL}
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pgc.User, pgc.Password, pgc.Host, pgc.Port, pgc.DB) // ?ssl=%s , ssl)
	var err error
	for attempt := 1; attempt <= int(bc.DB_MAX_RECONNECT_ATTEMPTS); attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
		db.Postgres, err = pgx.Connect(ctx, connString)
		if err != nil {
			cancel()
			return db, log.NewErrorWithTypef(ErrDatabaseConfig, "postgres connection string parsing: %s", err.Error())
		}
		err = db.Postgres.Ping(ctx)
		if err != nil {
			log.Logf(log.LevelInfo, "Connecting to database failed, attempt %d/%d", attempt, bc.DB_MAX_RECONNECT_ATTEMPTS)
			time.Sleep(bc.DB_RECONNECT_TIMEOUT_DURATION())
			cancel()
			continue
		}
		log.Logf(log.LevelInfo, "Postgres connection to database '%s' established.", pgc.DB)
		cancel()
		return db, log.ErrorNil()
	}
	return db, log.NewErrorWithType(ErrDatabaseConn, err.Error())
}
