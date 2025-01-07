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

func (db DB) DeleteRefreshToken(userID int, nonce string) *log.Error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()

	query := "DELETE FROM refresh_tokens WHERE user_id = $1 AND token_nonce = $2"
	res, err := db.Postgres.Exec(ctx, query, userID, nonce)
	if err != nil {
		return log.NewErrorWithTypef(ErrDatabaseDelete, "refresh token entry (rows affected: %d): %s", res.RowsAffected(), err.Error())
	}
	if res.RowsAffected() != 1 {
		return log.NewErrorWithTypef(ErrDatabaseDelete, "refresh token rows affected != 1 (instead got %d)", res.RowsAffected())
	}
	return log.ErrorNil()
}

func (db DB) VerifyRefreshTokenNonce(userID int, nonce string) (bool, *log.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	tx, err := db.Postgres.Begin(ctx)
	if err != nil {
		return false, log.NewErrorWithTypef(ErrDatabaseQuery, "unable to start db transaction: %s", err.Error())
	}
	_, errx := db.getTokenByUserIdAndNonce(ctx, tx, userID, nonce)
	if errx.IsType(ErrDatabaseNotFound) {
		return false, log.ErrorNil()
	}
	if !errx.IsNil() {
		return false, errx
	}
	err = tx.Commit(ctx)
	if err != nil {
		return false, log.NewErrorWithType(ErrDatabaseCommit, err.Error())
	}
	return true, log.ErrorNil()
}

func (db DB) UpdateRefreshTokenEntry(userId int, nonce string, newNonce string, expiresAt time.Time) *log.Error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	tx, err := db.Postgres.Begin(ctx)
	if err != nil {
		return log.NewErrorWithTypef(ErrDatabaseQuery, "unable to start db transaction: %s", err.Error())
	}
	defer tx.Rollback(context.Background())

	token, errx := db.getTokenByUserIdAndNonce(ctx, tx, userId, nonce)
	if !errx.IsNil() {
		return errx //.Extend("unable to get existing entry for token")
	}
	token.TokenNonce = newNonce
	token.ExpiresAt = expiresAt
	errx = db.updateToken(ctx, tx, token)
	if !errx.IsNil() {
		return errx //.Extend("unable to update refresh token entry")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return log.NewErrorWithType(ErrDatabaseCommit, err.Error())
	}
	return log.ErrorNil()

}

func (db DB) CreateRefreshTokenEntry(token tables.RefreshTokens) *log.Error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	query := "INSERT INTO refresh_tokens (user_id, token_nonce, reissue_count, expires_at) VALUES ($1, $2, $3, $4)"
	_, err := db.Postgres.Exec(ctx, query, token.UserID, token.TokenNonce, token.ReissueCount, token.ExpiresAt)
	if err != nil {
		return log.NewErrorWithTypef(ErrDatabaseInsert, "refresh token entry for user could not be created: %s", err.Error())
	}
	return log.ErrorNil()
}

func (db DB) getTokenByUserIdAndNonce(ctx context.Context, tx pgx.Tx, userID int, nonce string) (tables.RefreshTokens, *log.Error) {
	token := tables.RefreshTokens{}
	rows, err := tx.Query(ctx, "SELECT * FROM refresh_tokens WHERE user_id = $1 AND token_nonce = $2", userID, nonce)
	if errors.Is(err, pgx.ErrNoRows) {
		return token, log.NewErrorWithType(ErrDatabaseNotFound, "no refresh token found")
	}
	if err != nil {
		return token, log.NewErrorWithType(ErrDatabaseQuery, err.Error())
	}
	err = pgxscan.ScanOne(&token, rows)
	if err != nil {
		return token, log.NewErrorWithType(ErrDatabaseScan, err.Error())
	}
	return token, log.ErrorNil()
}

func (db DB) updateToken(ctx context.Context, tx pgx.Tx, token tables.RefreshTokens) *log.Error {
	query := "UPDATE refresh_tokens SET (token_nonce, reissue_count, updated_at, expires_at) = ($1, $2, $3, $4) WHERE id = $5"
	_, err := tx.Exec(ctx, query, token.TokenNonce, token.ReissueCount+1, token.UpdatedAt, token.ExpiresAt, token.ID)
	if err != nil {
		return log.NewErrorWithTypef(ErrDatabaseUpdate, "unable to update refresh token")
	}
	return log.ErrorNil()
}
