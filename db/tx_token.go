package db

import (
	"context"
	"time"

	"github.com/stytchauth/sqx"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/tables"
)

func VerifyRefreshTokenNonce(userID int, nonce string) (bool, *log.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	token, err := sqx.Read[tables.RefreshTokens](ctx).WithQueryable(Database).Select("*").From("refresh_tokens").Where(sqx.Eq{"user_id": userID, "token_nonce": nonce}).All()
	if err != nil {
		return false, log.NewErrorWithType(ErrDatabaseQuery, err.Error())
	}
	if len(token) < 1 {
		return false, log.ErrorNil()
	}
	if len(token) > 1 {
		return false, log.NewErrorWithTypef(ErrDatabaseQuery, "more than one token found for user (%d) and nonce (%s): %+v", userID, nonce, token)
	}
	return true, log.ErrorNil()
}
