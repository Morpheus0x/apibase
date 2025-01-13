package base

import (
	"gopkg.cc/apibase/db"
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

// setup pgx database connection, requires cleanup (on error, cleanup isn't required)
func (apiBase *ApiBase[T]) PostgresInit() (db.DB, error) {
	i := apiBase.registerCloseStage()

	database, err := db.PostgresInit(apiBase.Postgres, apiBase.BaseConfig, apiBase.CloseChain[i], apiBase.CloseChain[i+1])

	if err != nil {
		apiBase.CleanupOnError()
		return database, err
	}
	return database, nil
}

// start rest api server, is non-blocking, requires cleanup (on error, cleanup isn't required)
func (apiBase *ApiBase[T]) StartRest(api *web.ApiServer) error {
	i := apiBase.registerCloseStage()

	err := web_setup.StartRest(api, apiBase.ApiConfig.ApiBind, apiBase.CloseChain[i], apiBase.CloseChain[i+1])
	if err != nil {
		apiBase.CleanupOnError()
		return err
	}
	return nil
}
