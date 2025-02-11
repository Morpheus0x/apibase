package db

import (
	"context"
	"fmt"
	"time"

	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"

	"gopkg.cc/apibase/baseconfig"
	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/log"
)

// Initialize database connection, require shutdown and next channels for clean shutdown, these can be created using base.ApiBase[T].GetCloseStageChannels()
func PostgresInit(pgc PostgresConfig, bc *baseconfig.BaseConfig, shutdown chan struct{}, next chan struct{}) (DB, error) {
	db := DB{Kind: PostgreSQL, BaseConfig: bc}
	abort := make(chan struct{})

	go func() { // pgx shutdown
		select {
		case <-shutdown:
		case <-abort:
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), bc.TimeoutDatabaseShutdown)
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

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pgc.User, pgc.Password.GetSecret(), pgc.Host, pgc.Port, pgc.DB) // ?ssl=%s , ssl)
	var err error
	for attempt := 1; attempt <= int(bc.DatabaseMaxReconnectAttempts); attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), bc.TimeoutDatabaseConnect)
		db.Postgres, err = pgx.Connect(ctx, connString)
		if err != nil {
			log.Logf(log.LevelInfo, "Connecting to database failed, attempt %d/%d", attempt, bc.DatabaseMaxReconnectAttempts)
			time.Sleep(bc.DatabaseReconnectionTimeout)
			cancel()
			continue
		}

		pgxuuid.Register(db.Postgres.TypeMap())

		log.Logf(log.LevelInfo, "Postgres connection to database '%s' established.", pgc.DB)
		cancel()
		return db, nil
	}
	close(abort)
	return db, errx.WrapWithType(ErrDatabaseConn, err, "")
}
