package base

import (
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/web"
	"gopkg.cc/apibase/webtype"
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

// doesn't create go routines and therefore doesn't require cleanup
func (apiBase *ApiBase[T]) PostgresInit() *log.Error {
	var err *log.Error
	apiBase.ApiConfig.DB, err = db.PostgresInit(apiBase.Postgres, apiBase.BaseConfig)
	return err
}

// start rest api server, is non-blocking but requires cleanup
func (apiBase *ApiBase[T]) StartRest(api *webtype.ApiServer) *log.Error {
	i := apiBase.registerCloseStage()

	err := web.StartRest(api, apiBase.ApiConfig.ApiBind, apiBase.CloseChain[i], apiBase.CloseChain[i+1])
	if !err.IsNil() {
		apiBase.startupErrorCleanup()
		return err
	}
	return log.ErrorNil()
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
		// if this is the first entry in the channel close chain, close both shutdown and next channels and clear the array
		// all channel listeners should have been cleaned up by the setup function on error
		close(apiBase.CloseChain[0])
		close(apiBase.CloseChain[1])
		apiBase.CloseChain = nil
		return
	}
	// ccLength > 2
	// close and remove the last channel, therefore the new last element will be the "next" channel used to close subsequent go routines
	close(apiBase.CloseChain[ccLength-1])
	apiBase.CloseChain = apiBase.CloseChain[:ccLength-1]
}
