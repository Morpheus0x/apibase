package base

import (
	"context"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/web"
)

type ApiBase[T any] struct {
	Postgres   db.PostgresConfig `toml:"postgres"`
	SQLite     db.SQLiteConfig   `toml:"sqlite"`
	BaseConfig db.BaseConfig     `toml:"baseconfig"`
	ApiConfig  web.ApiConfig     `toml:"apiconfig"`

	Application T `toml:"application"`

	Interrupt  chan os.Signal
	CloseChain []chan struct{}
}

// initialize ApiBase struct without any additional application settings
func InitApiBase() *ApiBase[struct{}] {
	apiBase := &ApiBase[struct{}]{}

	// setup interrupt to catch sigint (Ctrl + C)
	apiBase.Interrupt = make(chan os.Signal, 1)
	signal.Notify(apiBase.Interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	return apiBase
}

// initialze ApiBase with custom config type used for decoding config file with additional application settings
func InitApiBaseCustom[T any]() *ApiBase[T] {
	apiBase := &ApiBase[T]{}

	// setup interrupt to catch sigint (Ctrl + C)
	apiBase.Interrupt = make(chan os.Signal, 1)
	signal.Notify(apiBase.Interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	return apiBase
}

// cleanup channel close chain array on error during registering a chain entry
func (apiBase *ApiBase[T]) CleanupOnError() {
	ccLength := len(apiBase.CloseChain)
	if ccLength < 1 {
		// nothing to clean up
		return
	}
	if ccLength < 2 {
		sleep := 500 * time.Millisecond // TODO: remove hardcoded timeout
		log.Log(log.LevelCritical, "close chain array only contains one channel, closing this channel. This should not happen!")
		close(apiBase.CloseChain[0])
		time.Sleep(sleep)
		return
	}
	if ccLength == 2 {
		// if this is the first entry in the channel close chain, close both shutdown and next channels and clear the array,
		// nothing should happen, since the errored stage shouldn't listen to this anymore
		close(apiBase.CloseChain[0])
		close(apiBase.CloseChain[1])
		apiBase.CloseChain = nil
		return
	}
	// ccLength > 2:
	// close shutdown channel of errored stage, nothing should happen, since the errored stage shouldn't listen to this anymore
	close(apiBase.CloseChain[0])
	// close shutdown channel of previous stage
	close(apiBase.CloseChain[1])
	// waiting for last channel in close chain to be closed
	timeout := 9 * time.Second // TODO: remove hardcoded timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	select {
	case <-apiBase.CloseChain[ccLength-1]:
		log.Log(log.LevelInfo, "all go routines closed")
	case <-ctx.Done():
		log.Logf(log.LevelNotice, "timeout for close chain exceeded, not all go routines exited in defined timeout (%s)", timeout.String())
	}

	// clear CloseChain array, since all channels have been closed
	apiBase.CloseChain = nil
}

func (apiBase *ApiBase[T]) Cleanup() {
	if len(apiBase.CloseChain) < 1 {
		return
	}
	if len(apiBase.CloseChain) < 2 {
		sleep := 500 * time.Millisecond // TODO: remove hardcoded timeout
		log.Logf(log.LevelCritical, "interrupt received but only one channel found, this should not happen, closing it and waiting %s for go routine to exit", sleep.String())
		close(apiBase.CloseChain[0])
		time.Sleep(sleep)
		return
	}

	close(apiBase.CloseChain[0])
	timeout := 9 * time.Second // TODO: remove hardcoded timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// waiting for last channel in close chain to be closed
	select {
	case <-apiBase.CloseChain[len(apiBase.CloseChain)-1]:
		log.Log(log.LevelInfo, "all go routines closed")
	case <-ctx.Done():
		log.Logf(log.LevelNotice, "timeout for close chain exceeded, not all go routines exited in defined timeout (%s)", timeout.String())
	}
}

func (apiBase *ApiBase[T]) WaitAndCleanup() error {
	if apiBase.Interrupt == nil {
		return errx.NewWithType(ErrApiBaseCleanup, "interrupt channel not initialized, make sure to initialize ApiBase struct correctly")
	}
	if len(apiBase.CloseChain) < 1 {
		log.Log(log.LevelWarning, "no close chain found and therefore no go routines can be closed, done")
		return nil
	}
	if len(apiBase.CloseChain) < 2 {
		return errx.NewWithType(ErrApiBaseCleanup, "only one channel in close chain, this should not happen, unable to close go routine(s)")
	}

	<-apiBase.Interrupt
	log.Log(log.LevelNotice, "interrupt received, closing go routines")
	close(apiBase.CloseChain[0])
	timeout := 9 * time.Second // TODO: remove hardcoded timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// waiting for last channel in close chain to be closed
	select {
	case <-apiBase.CloseChain[len(apiBase.CloseChain)-1]:
		log.Log(log.LevelInfo, "all go routines closed")
		return nil
	case <-ctx.Done():
		return errx.NewWithTypef(ErrApiBaseCleanup, "timeout for close chain exceeded, not all go routines exited in defined timeout (%s)", timeout.String())
	}
}

func (apiBase *ApiBase[T]) LoadToml(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return errx.WrapWithType(ErrTomlParsing, err, "unable to read toml file")
	}
	_, err = toml.DecodeFile(path, apiBase)
	if err != nil {
		return errx.WrapWithType(ErrTomlParsing, err, "unable to parse toml")
	}

	apiBase.AddMissingDefaults()

	return nil
}

func (apiBase *ApiBase[T]) AddMissingDefaults() {
	if reflect.ValueOf(apiBase.BaseConfig.DB_MAX_RECONNECT_ATTEMPTS).IsZero() {
		apiBase.BaseConfig.DB_MAX_RECONNECT_ATTEMPTS = DB_MAX_RECONNECT_ATTEMPTS
	}
	if reflect.ValueOf(apiBase.BaseConfig.DB_RECONNECT_TIMEOUT_SEC).IsZero() {
		apiBase.BaseConfig.DB_RECONNECT_TIMEOUT_SEC = DB_RECONNECT_TIMEOUT_SEC
	}
	if reflect.ValueOf(apiBase.BaseConfig.SQLITE_DATETIME_FORMAT).IsZero() {
		apiBase.BaseConfig.SQLITE_DATETIME_FORMAT = SQLITE_DATETIME_FORMAT
	}
	if reflect.ValueOf(apiBase.ApiConfig.DefaultOrgID).IsZero() {
		apiBase.ApiConfig.DefaultOrgID = DEFAULT_ORG_ID
	}
}

func (apiBase *ApiBase[T]) GetCustomConfigType() {
	log.Logf(log.LevelDebug, "Type: %s", reflect.TypeFor[T]().String())
}
