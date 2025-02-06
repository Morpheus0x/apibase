package db

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/table"
)

func (db DB) createOrg(org table.Organization, tx pgx.Tx, ctx context.Context) (table.Organization, error) {
	createdOrg := table.Organization{}
	query := "INSERT INTO organizations (name, description) VALUES ($1, $2) RETURNING id, name, description"
	rows, err := tx.Query(ctx, query, org.Name, org.Description)
	if err != nil {
		return createdOrg, errx.WrapWithTypef(ErrDatabaseInsert, err, "organization '%s' could not be created", org.Name)
	}
	err = pgxscan.ScanOne(&createdOrg, rows)
	if err != nil {
		return createdOrg, errx.WrapWithType(ErrDatabaseScan, err, "")
	}
	return createdOrg, nil
}
