package config

import (
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/web"
	"gopkg.cc/apibase/webtype"
)

// used to add a stage in the close channel chain, returns offset of channel to listen to,
// guarantees offset to be second to last element of chain
func (ab *ApiBase) RegisterCloseStage() uint {
	var offset uint
	if len(ab.CloseChain) < 1 {
		ab.CloseChain = append(ab.CloseChain, make(chan struct{}))
		offset = 0
	} else {
		offset = uint(len(ab.CloseChain) - 1)
	}
	ab.CloseChain = append(ab.CloseChain, make(chan struct{}))
	return offset
}

func (ab *ApiBase) PostgresInit() *log.Error {
	var err *log.Error
	ab.ApiConfig.DB, err = db.PostgresInit(ab.Postgres, ab.BaseConfig)
	return err
}

// start rest api server, is non-blocking
func (ab *ApiBase) StartRest(api *webtype.ApiServer) *log.Error {
	i := ab.RegisterCloseStage()

	err := web.StartRest(api, ab.ApiConfig.ApiBind, ab.CloseChain[i], ab.CloseChain[i+1])
	if !err.IsNil() {
		return err
	}
	return log.ErrorNil()
}
