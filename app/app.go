package app

import (
	"time"

	"gopkg.cc/apibase/db"
)

type ApiConfig struct {
	CORS []string
	DB   db.DB

	TokenSecret          string
	TokenAccessValidity  time.Duration
	TokenRefreshValidity time.Duration
}
