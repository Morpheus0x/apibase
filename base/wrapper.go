package base

import (
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/web"
	"gopkg.cc/apibase/webtype"
)

// used to add a stage in the close channel chain, returns offset of channel to listen to,
// guarantees offset to be second to last element of chain
func (apiBase *ApiBase[T]) RegisterCloseStage() uint {
	var offset uint
	if len(apiBase.CloseChain) < 1 {
		apiBase.CloseChain = append(apiBase.CloseChain, make(chan struct{}))
		offset = 0
	} else {
		offset = uint(len(apiBase.CloseChain) - 1)
	}
	apiBase.CloseChain = append(apiBase.CloseChain, make(chan struct{}))
	return offset
}

// doesn't create go routines and therefore doesn't require cleanup
func (apiBase *ApiBase[T]) PostgresInit() *log.Error {
	var err *log.Error
	apiBase.ApiConfig.DB, err = db.PostgresInit(apiBase.Postgres, apiBase.BaseConfig)
	return err
}

// start rest api server, is non-blocking but requires cleanup
func (apiBase *ApiBase[T]) StartRest(api *webtype.ApiServer) *log.Error {
	i := apiBase.RegisterCloseStage()

	err := web.StartRest(api, apiBase.ApiConfig.ApiBind, apiBase.CloseChain[i], apiBase.CloseChain[i+1])
	if !err.IsNil() {
		// cleanup channel close chain array
		if len(apiBase.CloseChain) == 2 {
			// if this is the first entry in the channel close chain remove both shutdown and next channels by clearing the array
			apiBase.CloseChain = nil
		} else if len(apiBase.CloseChain) > 2 {
			// remove the last channel, therefore the new last element will be the "next" channel used to close subsequent go routines
			apiBase.CloseChain = apiBase.CloseChain[:len(apiBase.CloseChain)-1]
		} else {
			log.Log(log.LevelCritical, "close chain array only contains one channel, this should not happen!")
		}
		return err
	}
	return log.ErrorNil()
}
