package db

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/table"
)

func (db DB) GetScheduledTasks(userId int) ([]table.ScheduledTask, error) {
	tasks := []table.ScheduledTask{}
	roles, err := db.GetUserRoles(userId)
	if err != nil {
		return tasks, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), db.BaseConfig.TimeoutDatabaseQuery)
	defer cancel()

	var noViewPermsForOrg []int
	for _, role := range roles {
		if !role.OrgView {
			noViewPermsForOrg = append(noViewPermsForOrg, role.OrgID)
			continue
		}
		orgTasks, err := db.getScheduledTasksForOrg(role.OrgID, ctx)
		if err != nil {
			log.Logf(log.LevelError, "unable to get tasks for org (id: %d): %s", role.OrgID, err.Error())
			continue
		}
		tasks = append(tasks, orgTasks...)
	}

	if len(noViewPermsForOrg) > 0 {
		return tasks, errx.NewWithTypef(errx.ErrSomeMinorOccurred, "User doesn't have view permission for organizations: %+v", noViewPermsForOrg)
	}
	return tasks, nil
}

func (db DB) getScheduledTasksForOrg(orgId int, ctx context.Context) ([]table.ScheduledTask, error) {
	tasks := []table.ScheduledTask{}
	query := "SELECT * FROM scheduled_tasks WHERE org_id = $1"
	rows, err := db.Postgres.Query(ctx, query, orgId)
	if err != nil {
		return tasks, errx.WrapWithTypef(ErrDatabaseQuery, err, "")
	}
	err = pgxscan.ScanAll(&tasks, rows)
	if err != nil {
		return tasks, errx.WrapWithTypef(ErrDatabaseScan, err, "")
	}
	// if no tasks are found, the empty array is returned
	return tasks, nil
}

func (db DB) GetAllScheduledTasks() ([]table.ScheduledTask, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.BaseConfig.TimeoutDatabaseQuery)
	defer cancel()
	tasks := []table.ScheduledTask{}
	rows, err := db.Postgres.Query(ctx, "SELECT * FROM scheduled_tasks")
	if err != nil {
		return tasks, errx.WrapWithType(ErrDatabaseQuery, err, "")
	}
	err = pgxscan.ScanAll(&tasks, rows)
	if err != nil {
		return tasks, errx.WrapWithType(ErrDatabaseScan, err, "")
	}
	if len(tasks) < 1 {
		return tasks, errx.NewWithType(ErrDatabaseNotFound, "no tasks found")
	}
	return tasks, nil
}

func (db DB) CreateScheduledTask(task table.ScheduledTask) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.BaseConfig.TimeoutDatabaseQuery)
	defer cancel()
	query := "INSERT INTO scheduled_tasks (task_id, org_id, start_date, interval, task_type, task_data) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err := db.Postgres.Exec(ctx, query, task.TaskID, task.OrgID, task.StartDate, task.Interval, task.TaskType, task.TaskData)
	if err != nil {
		return errx.WrapWithType(ErrDatabaseInsert, err, "scheduled task entry could not be created")
	}
	return nil
}

func (db DB) DeleteScheduledTask(taskId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.BaseConfig.TimeoutDatabaseQuery)
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

func (db DB) UpdateScheduledTask(task table.ScheduledTask) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.BaseConfig.TimeoutDatabaseQuery)
	defer cancel()
	query := "UPDATE scheduled_tasks SET (org_id, start_date, interval, task_type, task_data, updated_at) = ($1, $2, $3, $4, $5, $6) WHERE task_id = $7"
	_, err := db.Postgres.Exec(ctx, query, task.OrgID, task.StartDate, task.Interval, task.TaskType, task.TaskData, time.Now(), task.TaskID)
	if err != nil {
		return errx.WrapWithType(ErrDatabaseUpdate, err, "scheduled task could not be updated")
	}
	return nil
}
