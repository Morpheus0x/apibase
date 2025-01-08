package db

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"

	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/table"
)

func (db DB) CreateUserIfNotExist(user table.User, role table.UserRole) (table.User, *log.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	tx, err := db.Postgres.Begin(ctx)
	if err != nil {
		return user, log.NewErrorWithTypef(ErrDatabaseQuery, "unable to start db transaction: %s", err.Error())
	}
	defer tx.Rollback(context.Background())

	_, dbErr := db.getUserByEmail(user.Email, tx, ctx)
	if dbErr.IsNil() || (!dbErr.IsNil() && !dbErr.IsType(ErrDatabaseNotFound)) {
		// TODO: requires wrapped error with multiple error types for propper error handling
		return user, log.NewErrorWithTypef(ErrUserAlreadyExists, "user can't be created: %s", dbErr.String())
	}
	userFromDB, dbErr := db.createUser(user, tx, ctx)
	if !dbErr.IsNil() {
		return user, dbErr
	}
	role.UserID = userFromDB.ID
	dbErr = db.createUserRole(role, tx, ctx)
	if !dbErr.IsNil() {
		return userFromDB, dbErr.Extendf("unable to create role for user with email '%s'", user.Email)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return userFromDB, log.NewErrorWithType(ErrDatabaseCommit, err.Error())
	}
	return userFromDB, log.ErrorNil()
}

func (db DB) GetUserByID(id int) (table.User, *log.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	user := table.User{}
	// err := pgxscan.Select(ctx, Database, &user, "SELECT * FROM users WHERE id = $1", id)
	rows, err := db.Postgres.Query(ctx, "SELECT * FROM users WHERE id = $1", id)
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

func (db DB) GetUserByEmail(email string) (table.User, *log.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	tx, err := db.Postgres.Begin(ctx)
	if err != nil {
		return table.User{}, log.NewErrorWithTypef(ErrDatabaseQuery, "unable to start db transaction: %s", err.Error())
	}
	defer tx.Rollback(context.Background())

	user, dbErr := db.getUserByEmail(email, tx, ctx)
	if !dbErr.IsNil() {
		return user, dbErr
	}

	err = tx.Commit(ctx)
	if err != nil {
		return user, log.NewErrorWithType(ErrDatabaseCommit, err.Error())
	}
	return user, log.ErrorNil()
}

// unique user is defined by user.Email, also creates the default viewer role for the specified organization
func (db DB) GetOrCreateUser(user table.User, role table.UserRole) (table.User, *log.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	tx, err := db.Postgres.Begin(ctx)
	if err != nil {
		return user, log.NewErrorWithTypef(ErrDatabaseQuery, "unable to start db transaction: %s", err.Error())
	}
	defer tx.Rollback(context.Background())

	userFromDB, dbErr := db.getUserByEmail(user.Email, tx, ctx)
	if !dbErr.IsNil() && !dbErr.IsType(ErrDatabaseNotFound) {
		return user, dbErr
	}
	if dbErr.IsType(ErrDatabaseNotFound) {
		// TODO: validate, that user has non-empty string for NickName, Email
		userFromDB, dbErr = db.createUser(user, tx, ctx)
		if !dbErr.IsNil() {
			return user, dbErr
		}
		role.UserID = userFromDB.ID
		dbErr = db.createUserRole(role, tx, ctx)
		if !dbErr.IsNil() {
			return user, dbErr.Extendf("unable to create role for user with email '%s'", user.Email)
		}
		log.Logf(log.LevelDebug, "User created: %s (%s)", user.Name, user.Email)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return user, log.NewErrorWithType(ErrDatabaseCommit, err.Error())
	}
	return userFromDB, log.ErrorNil()
}

func (db DB) getUserByEmail(email string, tx pgx.Tx, ctx context.Context) (table.User, *log.Error) {
	user := table.User{}
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

func (db DB) createUser(user table.User, tx pgx.Tx, ctx context.Context) (table.User, *log.Error) {
	createdUser := table.User{}
	query := "INSERT INTO users (name, auth_provider, email, email_verified, password_hash, secrets_version, totp_secret, super_admin) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING (id, name, auth_provider, email, email_verified, password_hash, secrets_version, totp_secret, super_admin, created_at, updated_at)"
	rows, err := tx.Query(ctx, query, user.Name, user.AuthProvider, user.Email, user.EmailVerified, user.PasswordHash, user.SecretsVersion, user.TotpSecret, user.SuperAdmin)
	if err != nil {
		return createdUser, log.NewErrorWithTypef(ErrDatabaseInsert, "user (email: %s) could not be created: %s", user.Email, err.Error())
	}
	err = pgxscan.ScanOne(&createdUser, rows)
	if err != nil {
		return createdUser, log.NewErrorWithType(ErrDatabaseScan, err.Error())
	}
	return createdUser, log.ErrorNil()
}
