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

func CreateUser(user tables.Users) *log.Error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	tx, err := Database.Begin(ctx)
	if err != nil {
		return log.NewErrorWithTypef(ErrDatabaseQuery, "unable to start db transaction: %s", err.Error())
	}
	defer tx.Rollback(context.Background())

	errx := createUser(user, tx, ctx)
	if !errx.IsNil() {
		return errx
	}
	err = tx.Commit(ctx)
	if err != nil {
		return log.NewErrorWithType(ErrDatabaseCommit, err.Error())
	}
	return log.ErrorNil()
}

func GetUserByID(id int) (tables.Users, *log.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	user := tables.Users{}
	// err := pgxscan.Select(ctx, Database, &user, "SELECT * FROM users WHERE id = $1", id)
	rows, err := Database.Query(ctx, "SELECT * FROM users WHERE id = $1", id)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, log.NewErrorWithTypef(ErrDatabaseNotFound, "no user found for id '%d'", id)
	}
	if err != nil {
		return user, log.NewErrorWithType(ErrDatabaseQuery, err.Error())
	}
	err = pgxscan.ScanOne(&user, rows)
	if err != nil {
		return user, log.NewErrorWithType(ErrDatabaseScan, err.Error())
	}
	// err := pgxscan.NewScanner(row).Scan(&user)
	return user, log.ErrorNil()
}

// unique user is defined by user.Email
func GetOrCreateUser(user tables.Users) (tables.Users, *log.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	tx, err := Database.Begin(ctx)
	if err != nil {
		return user, log.NewErrorWithTypef(ErrDatabaseQuery, "unable to start db transaction: %s", err.Error())
	}
	defer tx.Rollback(context.Background())

	userFromDB, dbErr := getUserByEmail(user.Email, tx, ctx)
	if !dbErr.IsNil() && !dbErr.IsType(ErrDatabaseNotFound) {
		return user, dbErr
	}
	if dbErr.IsType(ErrDatabaseNotFound) {
		// TODO: validate, that user has non-empty string for NickName, Email
		dbErr = createUser(user, tx, ctx)
		if !dbErr.IsNil() {
			return user, dbErr
		}
		log.Logf(log.LevelDebug, "User created: %s (%s)", user.Name, user.Email)
		userFromDB, dbErr = getUserByEmail(user.Email, tx, ctx)
		if !dbErr.IsNil() {
			return user, dbErr.Extendf("unable to get just created user with email '%s'", user.Email)
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return user, log.NewErrorWithType(ErrDatabaseCommit, err.Error())
	}
	return userFromDB, log.ErrorNil()
}

func getUserByEmail(email string, tx pgx.Tx, ctx context.Context) (tables.Users, *log.Error) {
	user := tables.Users{}
	rows, err := tx.Query(ctx, "SELECT * FROM users WHERE email = $1", email)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, log.NewErrorWithTypef(ErrDatabaseNotFound, "no user found for email '%s'", email)
	}
	if err != nil {
		return user, log.NewErrorWithType(ErrDatabaseQuery, err.Error())
	}
	err = pgxscan.ScanOne(&user, rows)
	if err != nil {
		return user, log.NewErrorWithType(ErrDatabaseScan, err.Error())
	}
	return user, log.ErrorNil()
}

func createUser(user tables.Users, tx pgx.Tx, ctx context.Context) *log.Error {
	query := "INSERT INTO users (name, auth_provider, email, email_verified, password_hash, secrets_version, totp_secret, super_admin) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	_, err := tx.Exec(ctx, query, user.Name, user.AuthProvider, user.Email, user.EmailVerified, user.PasswordHash, user.SecretsVersion, user.TotpSecret, user.SuperAdmin)
	if err != nil {
		return log.NewErrorWithTypef(ErrDatabaseQuery, "user (email: %s) could not be created: %s", user.Email, err.Error())
	}
	return log.ErrorNil()
}
