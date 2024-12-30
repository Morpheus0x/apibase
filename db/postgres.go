package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"

	"gopkg.cc/apibase/log"
)

// require shutdown and next channels for clean shutdown, these can be created using base.ApiBase[T].GetCloseStageChannels()
func PostgresInit(pgc PostgresConfig, bc BaseConfig, shutdown chan struct{}, next chan struct{}) (DB, *log.Error) {
	db := DB{Kind: PostgreSQL}
	abort := make(chan struct{})

	go func() { // pgx shutdown
		select {
		case <-shutdown:
		case <-abort:
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // TODO: remove hardcoded timeout
		defer cancel()
		if db.Postgres != nil {
			err := db.Postgres.Close(ctx)
			if err != nil {
				log.Logf(log.LevelError, "unable to close postgres database connection: %s", err.Error())
			} else {
				log.Log(log.LevelNotice, "postgres database connection closed successful.")
			}
		}
		select {
		case <-next:
			log.Log(log.LevelError, "postgres next channel already closed, make sure channel staging is done correctly")
		default:
			close(next)
		}
	}()

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
	close(abort)
	return db, log.NewErrorWithType(ErrDatabaseConn, err.Error())
}
