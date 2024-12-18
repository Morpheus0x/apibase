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

func (apiBase *ApiBase[T]) PostgresInit() *log.Error {
	var err *log.Error
	apiBase.ApiConfig.DB, err = db.PostgresInit(apiBase.Postgres, apiBase.BaseConfig)
	return err
}

// start rest api server, is non-blocking
func (apiBase *ApiBase[T]) StartRest(api *webtype.ApiServer) *log.Error {
	i := apiBase.RegisterCloseStage()

	err := web.StartRest(api, apiBase.ApiConfig.ApiBind, apiBase.CloseChain[i], apiBase.CloseChain[i+1])
	if !err.IsNil() {
		return err
	}
	return log.ErrorNil()
}
