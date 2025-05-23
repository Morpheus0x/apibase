package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"

	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/table"
)

// If no roles are provided, a new organization is created for the user with admin role for user,
// otherwise CreateUserIfNotExist is run
func (db DB) CreateNewUserWithOrg(user table.User, roles ...table.UserRole) (table.User, error) {
	if len(roles) > 0 {
		// Don't create new org for user, instead assign user defined roles
		return db.CreateUserIfNotExist(user, roles...)
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.BaseConfig.TimeoutDatabaseQuery)
	defer cancel()
	tx, err := db.Postgres.Begin(ctx)
	if err != nil {
		return user, errx.WrapWithType(ErrDatabaseQuery, err, "unable to start db transaction")
	}
	defer tx.Rollback(context.Background())

	_, err = db.getUserByEmail(user.Email, tx, ctx)
	if err == nil || !errors.Is(err, ErrDatabaseNotFound) {
		return user, errx.WrapWithType(ErrUserAlreadyExists, err, "user can't be created")
	}
	userFromDB, err := db.createUser(user, tx, ctx)
	if err != nil {
		return user, err
	}
	org := table.Organization{
		Name:        fmt.Sprintf("userorg-%s", userFromDB.Name),
		Description: fmt.Sprintf("Default org for user '%s'", userFromDB.Name),
	}
	orgFromDB, err := db.createOrg(org, tx, ctx)
	if err != nil {
		return user, errx.WrapWithTypef(ErrOrgCreate, err, "for user (id: %d)", userFromDB.ID)
	}
	role := table.UserRole{
		UserID:   userFromDB.ID,
		OrgID:    orgFromDB.ID,
		OrgView:  true,
		OrgEdit:  true,
		OrgAdmin: true,
	}
	err = db.createUserRole(role, tx, ctx)
	if err != nil {
		return userFromDB, errx.Wrapf(err, "unable to create role for user with email '%s'", user.Email)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return userFromDB, errx.WrapWithType(ErrDatabaseCommit, err, "")
	}
	return userFromDB, nil
}

func (db DB) CreateUserIfNotExist(user table.User, roles ...table.UserRole) (table.User, error) {
	if len(roles) < 1 {
		return user, errx.NewWithType(ErrNoRoles, "CreateUserIfNotExist must have at least one role for the new user")
	}
	ctx, cancel := context.WithTimeout(context.Background(), db.BaseConfig.TimeoutDatabaseQuery)
	defer cancel()
	tx, err := db.Postgres.Begin(ctx)
	if err != nil {
		return user, errx.WrapWithType(ErrDatabaseQuery, err, "unable to start db transaction")
	}
	defer tx.Rollback(context.Background())

	_, err = db.getUserByEmail(user.Email, tx, ctx)
	if err == nil || !errors.Is(err, ErrDatabaseNotFound) {
		return user, errx.WrapWithType(ErrUserAlreadyExists, err, "user can't be created")
	}
	userFromDB, err := db.createUser(user, tx, ctx)
	if err != nil {
		return user, err
	}
	for _, role := range roles {
		role.UserID = userFromDB.ID
		err = db.createUserRole(role, tx, ctx)
		if err != nil {
			return userFromDB, errx.Wrapf(err, "unable to create role for user with email '%s'", user.Email)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return userFromDB, errx.WrapWithType(ErrDatabaseCommit, err, "")
	}
	return userFromDB, nil
}

func (db DB) GetUserByID(id int) (table.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.BaseConfig.TimeoutDatabaseQuery)
	defer cancel()
	user := table.User{}
	// err := pgxscan.Select(ctx, Database, &user, "SELECT * FROM users WHERE id = $1", id)
	rows, err := db.Postgres.Query(ctx, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return user, errx.WrapWithType(ErrDatabaseQuery, err, "")
	}
	err = pgxscan.ScanOne(&user, rows)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, errx.NewWithTypef(ErrDatabaseNotFound, "no user found for id '%d'", id)
	}
	if err != nil {
		return user, errx.WrapWithType(ErrDatabaseScan, err, "")
	}
	// err := pgxscan.NewScanner(row).Scan(&user)
	return user, nil
}

func (db DB) GetUserByEmail(email string) (table.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.BaseConfig.TimeoutDatabaseQuery)
	defer cancel()
	tx, err := db.Postgres.Begin(ctx)
	if err != nil {
		return table.User{}, errx.WrapWithType(ErrDatabaseQuery, err, "unable to start db transaction")
	}
	defer tx.Rollback(context.Background())

	user, err := db.getUserByEmail(email, tx, ctx)
	if err != nil {
		return user, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return user, errx.NewWithType(ErrDatabaseCommit, err.Error())
	}
	return user, nil
}

// unique user is defined by user.Email, also creates the default viewer role for the specified organization
func (db DB) GetOrCreateUser(user table.User, role table.UserRole) (table.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.BaseConfig.TimeoutDatabaseQuery)
	defer cancel()
	tx, err := db.Postgres.Begin(ctx)
	if err != nil {
		return user, errx.NewWithTypef(ErrDatabaseQuery, "unable to start db transaction: %s", err.Error())
	}
	defer tx.Rollback(context.Background())

	userFromDB, err := db.getUserByEmail(user.Email, tx, ctx)
	if err != nil && !errors.Is(err, ErrDatabaseNotFound) {
		return user, err
	}
	if errors.Is(err, ErrDatabaseNotFound) {
		// TODO: validate, that user has non-empty string for NickName, Email
		userFromDB, err = db.createUser(user, tx, ctx)
		if err != nil {
			return user, err
		}
		role.UserID = userFromDB.ID
		err = db.createUserRole(role, tx, ctx)
		if err != nil {
			return user, errx.Wrapf(err, "unable to create role for user with email '%s'", user.Email)
		}
		log.Logf(log.LevelDebug, "User created: %s (%s)", user.Name, user.Email)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return user, errx.WrapWithType(ErrDatabaseCommit, err, "")
	}
	return userFromDB, nil
}

func (db DB) getUserByEmail(email string, tx pgx.Tx, ctx context.Context) (table.User, error) {
	user := table.User{}
	rows, err := tx.Query(ctx, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		return user, errx.WrapWithType(ErrDatabaseQuery, err, "")
	}
	err = pgxscan.ScanOne(&user, rows)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, errx.NewWithTypef(ErrDatabaseNotFound, "no user found for email '%s'", email)
	}
	if err != nil {
		return user, errx.WrapWithType(ErrDatabaseScan, err, "")
	}
	return user, nil
}

func (db DB) createUser(user table.User, tx pgx.Tx, ctx context.Context) (table.User, error) {
	createdUser := table.User{}
	query := "INSERT INTO users (name, auth_provider, email, email_verified, password_hash, secrets_version, totp_secret, super_admin) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, name, auth_provider, email, email_verified, password_hash, secrets_version, totp_secret, super_admin, created_at, updated_at"
	rows, err := tx.Query(ctx, query, user.Name, user.AuthProvider, user.Email, user.EmailVerified, user.PasswordHash, user.SecretsVersion, user.TotpSecret, user.SuperAdmin)
	if err != nil {
		return createdUser, errx.WrapWithTypef(ErrDatabaseInsert, err, "user (email: %s) could not be created", user.Email)
	}
	err = pgxscan.ScanOne(&createdUser, rows)
	if err != nil {
		return createdUser, errx.WrapWithType(ErrDatabaseScan, err, "")
	}
	return createdUser, nil
}
