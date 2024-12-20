package db

import (
	"context"
	"time"

	"github.com/stytchauth/sqx"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/tables"
)

func GetUserByID(id int) (tables.Users, *log.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	user, dbErr := getUser(sqx.Eq{"id": id}, Database, ctx)
	if !dbErr.IsNil() {
		return user, dbErr
	}
	return user, log.ErrorNil()
}

// unique user is defined by user.Email
func GetOrCreateUser(user tables.Users) (tables.Users, *log.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	tx, err := Database.BeginTx(ctx, nil)
	if err != nil {
		return user, log.NewErrorWithTypef(ErrDatabaseQuery, "unable to start db transaction: %s", err.Error())
	}
	defer tx.Rollback()

	userFromDB, dbErr := getUser(sqx.Eq{"email": user.Email}, tx, ctx)
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
		userFromDB, dbErr = getUser(sqx.Eq{"email": user.Email}, tx, ctx)
		if !dbErr.IsNil() {
			return user, dbErr.Extendf("unable to get just created user with email '%s'", user.Email)
		}
	}
	err = tx.Commit()
	if err != nil {
		return user, log.NewErrorWithTypef(ErrDatabaseQuery, "unable to commit db transaction: %s", err.Error())
	}
	return userFromDB, log.ErrorNil()
}

func getUser(where sqx.Eq, queryable sqx.Queryable, ctx context.Context) (tables.Users, *log.Error) {
	user, err := sqx.Read[tables.Users](ctx).WithQueryable(queryable).Select("*").From("users").Where(where).All()
	if err != nil {
		return tables.Users{}, log.NewErrorWithType(ErrDatabaseQuery, err.Error())
	}
	if len(user) < 1 {
		return tables.Users{}, log.NewErrorWithTypef(ErrDatabaseNotFound, "no user found with specified where clause: %+v", where)
	}
	if len(user) > 1 {
		return tables.Users{}, log.NewErrorWithTypef(ErrDatabaseQuery, "more than one user found with specified where clause %+v", where)
	}
	return user[0], log.ErrorNil()
}

func createUser(user tables.Users, queryable sqx.Queryable, ctx context.Context) *log.Error {
	err := sqx.Write(ctx).WithQueryable(queryable).Insert("users").SetMap(sqx.ToSetMap(user)).Do()
	if err != nil {
		return log.NewErrorWithTypef(ErrDatabaseQuery, "user could not be created: %s", err.Error())
	}
	return log.ErrorNil()
}
