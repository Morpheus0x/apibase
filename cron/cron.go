package cron

import (
	"context"
	"sync"
	"time"

	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/log"
)

const (
	Daily   = 24 * time.Hour
	Weekly  = 7 * Daily
	Monthly = 31 * Daily
	Yearly  = 365 * Daily
)

type TaskFunc func() error

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

// Schedule new task, thread safe.
// interval must be at leas one minute and has some special behaviour if it has one of these specific values, it will run at the time specified in start:
// cron.Daily, cron.Weekly (at weekday of start), cron.Monthly (at day of month of start, day of start must not be later than the 28th), cron.Yearly (at start datetime)
func Schedule(id string, start time.Time, interval time.Duration, run TaskFunc) error {
	newTask := task{start: start, interval: interval, task: run}
	newTask.chanInit()

	activeTasks.Lock()
	if _, ok := activeTasks.tasks[id]; ok {
		activeTasks.Unlock()
		return errx.NewWithType(ErrTaskExists, "")
	}
	activeTasks.tasks[id] = newTask
	activeTasks.Unlock()

	go worker(id, newTask)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()

	select {
	case <-newTask.startupSuccess:
		close(newTask.startupFailed)

		return nil
	case err := <-newTask.startupFailed:
		activeTasks.Lock()
		delete(activeTasks.tasks, id)
		activeTasks.Unlock()

		return errx.WrapWithType(ErrTaskStartup, err, "")
	case <-ctx.Done():
		close(newTask.shutdown)

		activeTasks.Lock()
		delete(activeTasks.tasks, id)
		activeTasks.Unlock()

		return errx.NewWithType(ErrTaskStartTimeout, "")
	}
}

// Remove scheduled task, thread safe.
func Remove(id string) error {
	activeTasks.Lock()
	defer activeTasks.Unlock()
	if _, ok := activeTasks.tasks[id]; !ok {
		return errx.Newf("worker with id '%s' isn't currently registered", id)
	}
	close(activeTasks.tasks[id].shutdown)
	delete(activeTasks.tasks, id)
	return nil
}

// Shutdown all active tasks, thread safe
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

	startDiff := now.Sub(nextRun)
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
				log.Logf(log.LevelCritical, "worker task '%s' GetNextRun returned error, this should never happen!: %s", id, err.Error())
			}
			run = time.NewTimer(nextRun.Sub(now))
			log.Logf(log.LevelDebug, "worker '%s' will run the next time at %s (in %s)", id, nextRun.String(), nextRun.Sub(now).String())
		case <-job.shutdown:
			return
		}
	}
}
