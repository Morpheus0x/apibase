package db

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/table"
)

func (db DB) GetScheduledTasks() ([]table.ScheduledTask, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	tasks := []table.ScheduledTask{}
	rows, err := db.Postgres.Query(ctx, "SELECT * FROM scheduled_tasks")
	if errors.Is(err, pgx.ErrNoRows) {
		return tasks, errx.NewWithType(ErrDatabaseNotFound, "no tasks found")
	}
	if err != nil {
		return tasks, errx.WrapWithType(ErrDatabaseQuery, err, "")
	}
	err = pgxscan.ScanAll(&tasks, rows)
	if err != nil {
		return tasks, errx.WrapWithType(ErrDatabaseScan, err, "")
	}
	return tasks, nil
}

func (db DB) CreateScheduledTask(task table.ScheduledTask) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	query := "INSERT INTO scheduled_tasks (task_id, start_date, interval, description) VALUES ($1, $2, $3, $4)"
	_, err := db.Postgres.Exec(ctx, query, task.TaskID, task.StartDate, task.Interval, task.Description)
	if err != nil {
		return errx.WrapWithType(ErrDatabaseInsert, err, "scheduled task entry could not be created")
	}
	return nil
}

func (db DB) DeleteScheduledTask(taskId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()

	query := "DELETE FROM scheduled_tasks WHERE task_id = $1"
	res, err := db.Postgres.Exec(ctx, query, taskId)
	if err != nil {
		return errx.WrapWithTypef(ErrDatabaseDelete, err, "scheduled task token entry (rows affected: %d)", res.RowsAffected())
	}
	if res.RowsAffected() != 1 {
		return errx.NewWithTypef(ErrDatabaseDelete, "scheduled task rows affected != 1 (instead got %d)", res.RowsAffected())
	}
	return nil
}
