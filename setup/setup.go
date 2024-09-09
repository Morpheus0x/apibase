package setup

import (
	"gopkg.cc/apibase/app"
	"gopkg.cc/apibase/web"
)

func Rest(config app.ApiConfig) (*web.ApiServer, error) {
	err := validateDB(config.DB)
	if err != nil {
		return nil, err
	}
	apiServer := web.SetupRest(config)
	return apiServer, nil
}
