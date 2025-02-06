package cron

import (
	"context"
	"errors"
	"sync"
	"time"

	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/table"
	"gopkg.cc/apibase/web"
)

const (
	Daily   = 24 * time.Hour
	Weekly  = 7 * Daily
	Monthly = 31 * Daily
	Yearly  = 365 * Daily
)

type TaskFunc func() error

type Task struct {
	ID          string
	OrgID       int
	Start       time.Time
	Interval    time.Duration
	Description string
	Run         TaskFunc
}

type task struct {
	start          time.Time
	interval       time.Duration
	task           TaskFunc
	startupSuccess chan struct{}
	startupFailed  chan error
	shutdown       chan struct{}
	wasShutdown    chan struct{}
}

func (t *task) chanInit() {
	t.startupSuccess = make(chan struct{})
	t.startupFailed = make(chan error)
	t.shutdown = make(chan struct{})
	t.wasShutdown = make(chan struct{})
}

type tasks struct {
	tasks map[string]task // map[id]task
	sync.Mutex
}

var activeTasks tasks

func GetScheduledTasksForUser(api *web.ApiServer, userId int) ([]Task, error) {
	tasks := []Task{}
	tasksFromDB, err := api.DB.GetScheduledTasks(userId)
	if errors.Is(err, errx.ErrSomeMinorOccurred) {
		log.Log(log.LevelWarning, err.Error())
		// TODO: decide what else to do with view permission errors
		return tasks, nil
	}
	if err != nil {
		return tasks, err
	}
	for _, t := range tasksFromDB {
		tasks = append(tasks, Task{
			ID:          t.TaskID,
			OrgID:       t.OrgID,
			Start:       t.StartDate,
			Interval:    time.Duration(t.Interval),
			Description: t.Description,
			Run:         nil,
		})
	}
	return tasks, nil
}

// Get all tasks that were saved in the database,
// cron.StartScheduledTasks should be called after addint all corresponding cron.TaskFunc to the returned Task array
func GetScheduledTasksFromDB(api *web.ApiServer) ([]Task, error) {
	tasks := []Task{}
	tasksFromDB, err := api.DB.GetAllScheduledTasks()
	if err != nil {
		return tasks, err
	}
	for _, t := range tasksFromDB {
		tasks = append(tasks, Task{
			ID:          t.TaskID,
			OrgID:       t.OrgID,
			Start:       t.StartDate,
			Interval:    time.Duration(t.Interval),
			Description: t.Description,
			Run:         nil,
		})
	}
	return tasks, nil
}

// Starts all scheduled tasks that were saved in the database,
// assign cron.TaskFunc to all Task array entries returned by cron.GetScheduledTasksFromDB before running this function,
// if any error occurs, any scheduled tasks will be shut down
func StartScheduledTasks(tasks []Task) error {
	// verify that all tasks have a function to run
	for _, t := range tasks {
		if t.Run == nil {
			return errx.NewWithTypef(ErrTaskNoRunFunc, "task: %s", t.ID)
		}
	}
	// schedule tasks
	startErrors := make(map[string]error)
	for _, t := range tasks {
		err := Schedule(t)
		if err != nil {
			startErrors[t.ID] = err
		}
	}
	if len(startErrors) > 0 {
		for id, err := range startErrors {
			log.Logf(log.LevelError, "Unable to start scheduled task(%s): %s", id, err.Error())
		}
		err := Shutdown()
		return errx.Wrap(err, "Starting scheduled tasks failed, see previous log output for detailed errors, any successfully scheduled tasks have been shut down")
	}
	return nil
}

// Schedule new task and save to database, thread safe.
// interval must be at leas one minute and has some special behaviour if it has one of these specific values, it will run at the time specified in start:
// cron.Daily, cron.Weekly (at weekday of start), cron.Monthly (at day of month of start, day of start must not be later than the 28th), cron.Yearly (at start datetime)
func ScheduleAndSaveToDB(api *web.ApiServer, t Task) error {
	err := Schedule(t)
	if err != nil {
		return err
	}
	err = api.DB.CreateScheduledTask(table.ScheduledTask{TaskID: t.ID, OrgID: t.OrgID, StartDate: t.Start, Interval: table.Duration(t.Interval), Description: t.Description})
	if err != nil {
		return errx.WrapWithType(ErrTaskDatabaseSave, err, "task successfully scheduled but not saved in database")
	}
	return nil
}

// Schedule new task, thread safe.
// interval must be at leas one minute and has some special behaviour if it has one of these specific values, it will run at the time specified in start:
// cron.Daily, cron.Weekly (at weekday of start), cron.Monthly (at day of month of start, day of start must not be later than the 28th), cron.Yearly (at start datetime)
func Schedule(t Task) error {
	if t.Run == nil {
		return errx.NewWithType(ErrTaskNoRunFunc, "")
	}
	newTask := task{start: t.Start, interval: t.Interval, task: t.Run}
	newTask.chanInit()

	activeTasks.Lock()
	if _, ok := activeTasks.tasks[t.ID]; ok {
		activeTasks.Unlock()
		return errx.NewWithType(ErrTaskExists, "")
	}
	activeTasks.tasks[t.ID] = newTask
	activeTasks.Unlock()

	go worker(t.ID, newTask)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()

	select {
	case <-newTask.startupSuccess:
		close(newTask.startupFailed)

		return nil
	case err := <-newTask.startupFailed:
		activeTasks.Lock()
		delete(activeTasks.tasks, t.ID)
		activeTasks.Unlock()

		return errx.WrapWithType(ErrTaskStartup, err, "")
	case <-ctx.Done():
		close(newTask.shutdown)

		activeTasks.Lock()
		delete(activeTasks.tasks, t.ID)
		activeTasks.Unlock()

		return errx.NewWithType(ErrTaskStartTimeout, "")
	}
}

// Remove scheduled task and also from database, thread safe.
func RemoveAlsoFromDB(api *web.ApiServer, id string) error {
	err := Remove(id)
	dbErr := api.DB.DeleteScheduledTask(id)
	if dbErr != nil {
		return errx.WrapWithType(ErrTaskDatabaseDelete, err, "")
	}
	if err != nil {
		return errx.Wrap(err, "The scheduled task was successfully removed from database, however")
	}
	return nil
}

// Remove scheduled task, thread safe.
func Remove(id string) error {
	activeTasks.Lock()
	defer activeTasks.Unlock()
	if _, ok := activeTasks.tasks[id]; !ok {
		return errx.NewWithTypef(ErrTaskRemove, "worker with id '%s' isn't currently registered", id)
	}
	close(activeTasks.tasks[id].shutdown)
	delete(activeTasks.tasks, id)
	return nil
}

// Shutdown all active tasks, doesn't remove scheduled tasks from database, thread safe
func Shutdown() error {
	timeout := 60 * time.Second // TODO: remove hardcoded timeout
	timeoutExceededCount := 0

	activeTasks.Lock()
	for id, t := range activeTasks.tasks {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		close(t.shutdown)
		select {
		case <-t.wasShutdown:
			// nothing to do
		case <-ctx.Done():
			timeoutExceededCount++
			log.Logf(log.LevelWarning, "unable to shutdown task (id: %s), timeout (%s) exceeded", id, timeout.String())
		}
		cancel()
	}
	activeTasks.Unlock()

	if timeoutExceededCount > 0 {
		return errx.Newf("%d scheduled tasks couldn't be shutdown within their timeout of %s each", timeoutExceededCount, timeout.String())
	}
	return nil
}

// is used internally to get the exact time when the task should run next,
// can be used externally to recompute the date for long existing tasks
func GetNextRun(now time.Time, start time.Time, interval time.Duration) (time.Time, error) {
	var nextRun time.Time
	if interval < time.Minute {
		return nextRun, errx.New("task interval must be at least one minute")
	}
	if start.After(now) {
		return start, nil
	}

	startDiff := now.Sub(start)
	pastIntervals := startDiff / interval
	pastFraction := startDiff % interval

	nextRun = start.Add(pastIntervals * interval)
	if pastFraction != 0 {
		nextRun = nextRun.Add(interval)
	}

	if interval != Daily && interval != Weekly && interval != Monthly && interval != Yearly {
		return nextRun, nil
	}
	if interval == Monthly && start.Day() > 28 {
		return nextRun, errx.New("day of month must not be later than the 28th when using special monthly interval")
	}

	// set specific run time for special intervals
	return time.Date(
		nextRun.Year(),
		nextRun.Month(),
		nextRun.Day(),
		start.Hour(),
		start.Minute(),
		start.Second(),
		start.Nanosecond(),
		time.Local,
	), nil
}

func worker(id string, job task) {
	now := time.Now()
	var nextRun time.Time
	var err error
	nextRun, err = GetNextRun(now, job.start, job.interval)
	if err != nil {
		job.startupFailed <- err
		return
	}
	run := time.NewTimer(nextRun.Sub(now))
	log.Logf(log.LevelDebug, "worker '%s' will run for the first time at %s (in %s)", id, nextRun.String(), nextRun.Sub(now).String())
	close(job.startupSuccess)
	for {
		select {
		case <-run.C:
			err := job.task()
			if err != nil {
				log.Logf(log.LevelError, "worker task '%s' returned error: %s", id, err.Error())
			}
			now = time.Now()
			nextRun, err = GetNextRun(now, job.start, job.interval)
			if err != nil {
				log.Logf(log.LevelCritical, "worker task '%s' shut down because GetNextRun returned error, this should never happen!: %s", id, err.Error())
				return
			}
			run = time.NewTimer(nextRun.Sub(now))
			log.Logf(log.LevelDebug, "worker '%s' will run the next time at %s (in %s)", id, nextRun.String(), nextRun.Sub(now).String())
		case <-job.shutdown:
			return
		}
	}
}
