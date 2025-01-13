package base

import (
	"context"
	"time"

	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/web"
	"gopkg.cc/apibase/web_setup"
)

// used to add a stage in the close channel chain, returns offset of channel to listen to,
// will always be 0, since cleanup is done in reverse order
func (apiBase *ApiBase[T]) registerCloseStage() uint {
	if len(apiBase.CloseChain) < 1 {
		// if CloseChain is an empty array, also create the "next" channel
		apiBase.CloseChain = append(apiBase.CloseChain, make(chan struct{}))
	}
	apiBase.CloseChain = append([]chan struct{}{make(chan struct{})}, apiBase.CloseChain...)
	// since the close chain array should cleanup everything in reverse order,
	// the new close channel will always be prepended.
	// Therefore this function always returns index 0
	return 0 // TODO: maybe don't return index at all
}

// returns shutdown and next channels which can be used for own application go routines
func (apiBase *ApiBase[T]) GetCloseStageChannels() (shutdown chan struct{}, next chan struct{}) {
	i := apiBase.registerCloseStage()
	return apiBase.CloseChain[i], apiBase.CloseChain[i+1]
}

// setup pgx database connection, requires cleanup
func (apiBase *ApiBase[T]) PostgresInit() (db.DB, error) {
	i := apiBase.registerCloseStage()

	database, err := db.PostgresInit(apiBase.Postgres, apiBase.BaseConfig, apiBase.CloseChain[i], apiBase.CloseChain[i+1])

	if err != nil {
		apiBase.startupErrorCleanup()
		return database, err
	}
	return database, nil
}

// start rest api server, is non-blocking but requires cleanup
func (apiBase *ApiBase[T]) StartRest(api *web.ApiServer) error {
	i := apiBase.registerCloseStage()

	err := web_setup.StartRest(api, apiBase.ApiConfig.ApiBind, apiBase.CloseChain[i], apiBase.CloseChain[i+1])
	if err != nil {
		apiBase.startupErrorCleanup()
		return err
	}
	return nil
}

// cleanup channel close chain array
func (apiBase *ApiBase[T]) startupErrorCleanup() {
	ccLength := len(apiBase.CloseChain)
	if ccLength < 1 {
		// nothing to clean up
		return
	}
	if ccLength < 2 {
		log.Log(log.LevelCritical, "close chain array only contains one channel, this should not happen!")
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
