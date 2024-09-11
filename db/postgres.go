package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.cc/apibase/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func PostgresInit(pgc PostgresConfig, bc config.BaseConfig) (DB, error) {
	ssl := ""
	if pgc.SSLMode {
		ssl = "enable"
	} else {
		ssl = "disable"
	}
	conn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", pgc.Host, pgc.Port, pgc.User, pgc.Password, pgc.DB, ssl)
	var dbErr error
	for attempt := 1; attempt <= int(bc.DB_MAX_RECONNECT_ATTEMPTS); attempt++ {
		// TODO: use own logger
		var db *gorm.DB
		db, dbErr = gorm.Open(postgres.Open(conn), &gorm.Config{Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: time.Second * 2,
			},
		)})
		if dbErr != nil {
			// TODO: use logger instead of printf
			fmt.Printf("Connecting to database failed, attempt %d/%d", attempt, bc.DB_MAX_RECONNECT_ATTEMPTS)
			time.Sleep(bc.DB_RECONNECT_TIMEOUT_DURATION())
			continue
		}
		return DB{Kind: PostgreSQL, Gorm: db}, nil
	}
	fmt.Printf("Unable to establish PG database connection: %v", dbErr.Error())
	return DB{}, dbErr
}
