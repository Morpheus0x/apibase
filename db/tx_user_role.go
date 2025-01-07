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

func (db DB) GetUserRoles(userID int) ([]tables.UserRoles, *log.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	roles := []tables.UserRoles{}
	// err := pgxscan.Select(ctx, Database, &user, "SELECT * FROM users WHERE id = $1", id)
	rows, err := db.Postgres.Query(ctx, "SELECT * FROM user_roles WHERE id = $1", userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return roles, log.NewErrorWithTypef(ErrDatabaseNotFound, "no roles found for user (id: %d)", userID)
	}
	if err != nil {
		return roles, log.NewErrorWithType(ErrDatabaseQuery, err.Error())
	}
	err = pgxscan.ScanAll(&roles, rows)
	if err != nil {
		return roles, log.NewErrorWithType(ErrDatabaseScan, err.Error())
	}
	// err := pgxscan.NewScanner(row).Scan(&user)
	return roles, log.ErrorNil()
}

func (db DB) getUserRole(userID int, orgID int, tx pgx.Tx, ctx context.Context) (tables.UserRoles, *log.Error) {
	role := tables.UserRoles{}
	rows, err := tx.Query(ctx, "SELECT * FROM user_roles WHERE user_id = $1 AND org_id = $2", userID, orgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return role, log.NewErrorWithTypef(ErrDatabaseNotFound, "no role found for user (id: %d) and org (id: %d)", userID, orgID)
	}
	if err != nil {
		return role, log.NewErrorWithType(ErrDatabaseQuery, err.Error())
	}
	err = pgxscan.ScanOne(&role, rows)
	if err != nil {
		return role, log.NewErrorWithType(ErrDatabaseScan, err.Error())
	}
	return role, log.ErrorNil()
}

func (db DB) createUserRole(role tables.UserRoles, tx pgx.Tx, ctx context.Context) *log.Error {
	query := "INSERT INTO user_roles (user_id, org_id, org_view, org_edit, org_admin) VALUES ($1, $2, $3, $4, $5)"
	_, err := tx.Exec(ctx, query, role.UserID, role.OrgID, role.OrgView, role.OrgEdit, role.OrgAdmin)
	if err != nil {
		return log.NewErrorWithTypef(ErrDatabaseInsert, "role for user (id: %d) could not be created: %s", role.UserID, err.Error())
	}
	return log.ErrorNil()
}
