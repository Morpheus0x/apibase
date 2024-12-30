package db

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/tables"
)

func VerifyRefreshTokenNonce(userID int, nonce string) (bool, *log.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	token := tables.RefreshTokens{}
	rows, err := Database.Query(ctx, "SELECT * FROM refresh_tokens WHERE user_id = $1 AND token_nonce = $2", userID, nonce)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, log.ErrorNil()
	}
	if err != nil {
		return false, log.NewErrorWithType(ErrDatabaseQuery, err.Error())
	}
	// TODO: maybe remove scanning of struct, since it isn't needed for this func return
	err = pgxscan.ScanOne(&token, rows)
	if err != nil {
		return false, log.NewErrorWithType(ErrDatabaseScan, err.Error())
	}
	return true, log.ErrorNil()
}
