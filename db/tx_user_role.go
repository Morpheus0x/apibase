package db

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/table"
)

func (db DB) GetUserRoles(userID int) ([]table.UserRole, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.BaseConfig.TimeoutDatabaseQuery)
	defer cancel()
	roles := []table.UserRole{}
	// err := pgxscan.Select(ctx, Database, &user, "SELECT * FROM users WHERE id = $1", id)
	rows, err := db.Postgres.Query(ctx, "SELECT * FROM user_roles WHERE user_id = $1", userID)
	if err != nil {
		return roles, errx.WrapWithType(ErrDatabaseQuery, err, "")
	}
	err = pgxscan.ScanAll(&roles, rows)
	if err != nil {
		return roles, errx.WrapWithType(ErrDatabaseScan, err, "")
	}
	if len(roles) < 1 {
		return roles, errx.NewWithTypef(ErrDatabaseNotFound, "no roles found for user (id: %d)", userID)
	}
	// err := pgxscan.NewScanner(row).Scan(&user)
	return roles, nil
}

func (db DB) getUserRole(userID int, orgID int, tx pgx.Tx, ctx context.Context) (table.UserRole, error) {
	role := table.UserRole{}
	rows, err := tx.Query(ctx, "SELECT * FROM user_roles WHERE user_id = $1 AND org_id = $2", userID, orgID)
	if err != nil {
		return role, errx.WrapWithType(ErrDatabaseQuery, err, "")
	}
	err = pgxscan.ScanOne(&role, rows)
	if errors.Is(err, pgx.ErrNoRows) {
		return role, errx.NewWithTypef(ErrDatabaseNotFound, "no role found for user (id: %d) and org (id: %d)", userID, orgID)
	}
	if err != nil {
		return role, errx.WrapWithType(ErrDatabaseScan, err, "")
	}
	return role, nil
}

func (db DB) createUserRole(role table.UserRole, tx pgx.Tx, ctx context.Context) error {
	query := "INSERT INTO user_roles (user_id, org_id, org_view, org_edit, org_admin) VALUES ($1, $2, $3, $4, $5)"
	_, err := tx.Exec(ctx, query, role.UserID, role.OrgID, role.OrgView, role.OrgEdit, role.OrgAdmin)
	if err != nil {
		return errx.WrapWithTypef(ErrDatabaseInsert, err, "role for user (id: %d) could not be created", role.UserID)
	}
	return nil
}
