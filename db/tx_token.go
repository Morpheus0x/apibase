package db

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"gopkg.cc/apibase/errx"
	h "gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/table"
)

func (db DB) DeleteRefreshToken(userID int, nonce h.SecretString) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()

	query := "DELETE FROM refresh_tokens WHERE user_id = $1 AND token_nonce = $2"
	// TODO: test that nonce.GetSecret() isn't necessary
	res, err := db.Postgres.Exec(ctx, query, userID, nonce)
	if err != nil {
		return errx.WrapWithTypef(ErrDatabaseDelete, err, "refresh token entry (rows affected: %d)", res.RowsAffected())
	}
	if res.RowsAffected() != 1 {
		return errx.NewWithTypef(ErrDatabaseDelete, "refresh token rows affected != 1 (instead got %d)", res.RowsAffected())
	}
	return nil
}

func (db DB) VerifyRefreshTokenNonce(userID int, nonce h.SecretString) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	tx, err := db.Postgres.Begin(ctx)
	if err != nil {
		return false, errx.WrapWithType(ErrDatabaseQuery, err, "unable to start db transaction: %s")
	}
	_, err = db.getTokenByUserIdAndNonce(ctx, tx, userID, nonce)
	if e, ok := err.(*errx.BaseError); ok {
		if e.Is(ErrDatabaseNotFound) {
			return false, nil
		}
	}
	if err != nil {
		return false, err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return false, errx.NewWithType(ErrDatabaseCommit, err.Error())
	}
	return true, nil
}

func (db DB) UpdateRefreshTokenEntry(userId int, nonce h.SecretString, newNonce h.SecretString, expiresAt time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	tx, err := db.Postgres.Begin(ctx)
	if err != nil {
		return errx.WrapWithType(ErrDatabaseQuery, err, "unable to start db transaction")
	}
	defer tx.Rollback(context.Background())

	token, err := db.getTokenByUserIdAndNonce(ctx, tx, userId, nonce)
	if err != nil {
		return errx.Wrap(err, "unable to get existing entry for token")
	}
	token.TokenNonce = newNonce
	token.ExpiresAt = expiresAt
	err = db.updateToken(ctx, tx, token)
	if err != nil {
		return errx.Wrap(err, "unable to update refresh token entry")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errx.WrapWithType(ErrDatabaseCommit, err, "")
	}
	return nil

}

func (db DB) CreateRefreshTokenEntry(token table.RefreshToken) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	query := "INSERT INTO refresh_tokens (user_id, token_nonce, reissue_count, expires_at) VALUES ($1, $2, $3, $4)"
	_, err := db.Postgres.Exec(ctx, query, token.UserID, token.TokenNonce, token.ReissueCount, token.ExpiresAt)
	if err != nil {
		return errx.WrapWithType(ErrDatabaseInsert, err, "refresh token entry for user could not be created")
	}
	return nil
}

func (db DB) getTokenByUserIdAndNonce(ctx context.Context, tx pgx.Tx, userID int, nonce h.SecretString) (table.RefreshToken, error) {
	token := table.RefreshToken{}
	// TODO: test that nonce.GetSecret() isn't necessary
	rows, err := tx.Query(ctx, "SELECT * FROM refresh_tokens WHERE user_id = $1 AND token_nonce = $2", userID, nonce)
	if errors.Is(err, pgx.ErrNoRows) {
		return token, errx.NewWithType(ErrDatabaseNotFound, "no refresh token found")
	}
	if err != nil {
		return token, errx.WrapWithType(ErrDatabaseQuery, err, "")
	}
	err = pgxscan.ScanOne(&token, rows)
	if err != nil {
		return token, errx.WrapWithType(ErrDatabaseScan, err, "")
	}
	return token, nil
}

func (db DB) updateToken(ctx context.Context, tx pgx.Tx, token table.RefreshToken) error {
	query := "UPDATE refresh_tokens SET (token_nonce, reissue_count, updated_at, expires_at) = ($1, $2, $3, $4) WHERE id = $5"
	// TODO: test that token.TokenNonce.GetSecret() isn't necessary
	_, err := tx.Exec(ctx, query, token.TokenNonce, token.ReissueCount+1, token.UpdatedAt, token.ExpiresAt, token.ID)
	if err != nil {
		return errx.NewWithTypef(ErrDatabaseUpdate, "unable to update refresh token")
	}
	return nil
}
