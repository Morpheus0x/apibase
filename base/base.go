package base

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.cc/apibase/cmd"
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/email"
	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/web"
)

type ApiBase[T any] struct {
	Postgres   db.PostgresConfig `toml:"postgres"`
	SQLite     db.SQLiteConfig   `toml:"sqlite"`
	BaseConfig db.BaseConfig     `toml:"baseconfig"`
	ApiConfig  web.ApiConfig     `toml:"apiconfig"`

	Email     map[string]email.EmailConfig   `toml:"email"`
	EmailTmpl map[string]email.EmailTemplate `toml:"email_template"`

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
	timeout := 3 * time.Second // TODO: remove hardcoded timeout
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

// This shouldn't be used in most cases, WaitAndCleanup() is preferred
func (apiBase *ApiBase[T]) Cleanup() error {
	if len(apiBase.CloseChain) < 1 {
		log.Log(log.LevelWarning, "no close chain found and therefore no go routines can be closed")
		return nil
	}
	if len(apiBase.CloseChain) < 2 {
		sleep := 500 * time.Millisecond // TODO: remove hardcoded timeout
		log.Logf(log.LevelCritical, "interrupt received but only one channel found, this should not happen, closing it and waiting %s for go routine to exit", sleep.String())
		close(apiBase.CloseChain[0])
		time.Sleep(sleep)
		return nil
	}

	close(apiBase.CloseChain[0])
	timeout := 3 * time.Second // TODO: remove hardcoded timeout
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

// Wait for interrupt and cleanup go routines once received (if any)
func (apiBase *ApiBase[T]) WaitAndCleanup() error {
	if apiBase.Interrupt == nil {
		err := apiBase.Cleanup()
		return errx.WrapWithType(ErrApiBaseCleanup, err, "interrupt channel not initialized, make sure to initialize ApiBase struct correctly, cleaning up")
	}
	<-apiBase.Interrupt
	log.Log(log.LevelNotice, "interrupt received, closing go routines")
	return apiBase.Cleanup()
}

func (apiBase *ApiBase[T]) LoadToml(settings cmd.Settings) error {
	if stat, err := os.Stat(settings.ConfigFile); err != nil || stat.IsDir() {
		return errx.WrapWithType(ErrTomlParsing, err, "unable to read toml file")
	}
	if _, err := toml.DecodeFile(settings.ConfigFile, apiBase); err != nil {
		return errx.WrapWithType(ErrTomlParsing, err, "unable to parse toml")
	}
	if settings.ApiRoot != "" {
		if u, err := url.ParseRequestURI(settings.ApiRoot); err == nil && u.Scheme != "" && u.Host != "" {
			apiBase.ApiConfig.ApiRoot = web.RootOptions{
				Kind:   "proxy",
				Target: settings.ApiRoot,
			}
		} else {
			if stat, err := os.Stat(settings.ApiRoot); err != nil || !stat.IsDir() {
				return errx.WrapWithTypef(ErrApiRootParsing, err, "string doesn't contain path or uri: %s", settings.ApiRoot)
			}
			apiBase.ApiConfig.ApiRoot = web.RootOptions{
				Kind:   "static",
				Target: settings.ApiRoot,
			}
		}
	}
	if err := apiBase.ParseEmailConfig(); err != nil {
		return err
	}

	apiBase.AddMissingDefaults()

	return nil
}

func (apiBase *ApiBase[T]) ParseEmailConfig() error {
	for key, ec := range apiBase.Email {
		if ec.ExchangeURI == "" && (ec.Host == "" || ec.Port == 0) {
			return errx.NewWithType(ErrEmailParsing, "exchange_uri or host+port is required")
		}
		if ec.ExchangeURI != "" && (ec.Host != "" || ec.Port != 0) {
			log.Logf(log.LevelWarning, "Parsing config: exchange_uri is set alongside host/port in [email.%s], either exchange_uri or host+port should be set, assuming email config for exchange server", key)
		}
		if ec.ExchangeURI != "" {
			ec.Sender = email.Exchange
			return errx.NewWithType(errx.ErrNotImplemented, "sending email via microsoft exchange")
		} else {
			ec.Sender = email.SMTP
		}
		apiBase.Email[key] = ec
	}
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
}

func (apiBase *ApiBase[T]) GetCustomConfigType() {
	log.Logf(log.LevelDebug, "Type: %s", reflect.TypeFor[T]().String())
}
