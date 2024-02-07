package apibase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func randToken(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

type SetupData struct {
	PostgresUri      string // e.g. postgres://localhost:5432/gopsql?sslmode=disable
	PostgresConfig   *pgx.ConnConfig
	PostgresPassword string
	Bind             string // e.g. 0.0.0.0:9090
}

type Ident string

type Integration struct {
	Id Ident
}

type NewApi struct {
	Integration
	Name string
}

type Api struct {
	Uri string
	// AuthType APIAuthType
	Auth string
}

type Data struct {
	DataUri string
}

type ApiId string
type DataId string

type Application struct {
	ApiIntegration  map[ApiId]Api
	DataIntegration map[DataId]Data
}

func (a *Application) RegisterApiIntegration(uri string, auth string) ApiId { // id string,
	// _, ok := a.ApiIntegration[id]
	// if !ok {
	// 	panic(fmt.Errorf("integration already exists"))
	// }
	id, err := randToken(16)
	if err != nil {
		panic(fmt.Errorf(""))
	}
	a.ApiIntegration[ApiId(id)] = Api{Uri: uri, Auth: auth}
	return ApiId(id)
}

type Request struct {
	headers map[string]string
	body    string
}

type Response struct {
	body string
}

type Handler func(Request, ...Integration) Response

func (a *Application) RegisterEndpoint(path string, handler Handler) {

}

type API struct {
	Integration
	Uri  string
	Auth string
}

func createApp() Application {
	return Application{ApiIntegration: make(map[ApiId]Api), DataIntegration: make(map[DataId]Data)}
}

func myHandler(req Request, integr ...Integration) Response {
	return Response{body: ""}
}

func startup() {
	app := createApp()
	myapi := app.RegisterApiIntegration("https://localhost:99/", "none")
	app.RegisterEndpoint("/test", myHandler)
}

func test(data SetupData) {
	config := &pgx.ConnConfig{}
	var err error
	if data.PostgresUri != "" {
		config, err = pgx.ParseConfig(data.PostgresUri)
		if err != nil {
			panic(err) // TODO: improve error handling
		}
	} else {
		config = data.PostgresConfig
	}
	// conn, err = pgx.Connect("postgres://localhost:5432/gopsql?sslmode=disable")
	conn, err := pgx.ConnectConfig(context.Background(), config)
	if err != nil {
		panic(err)
	}
	conn.Close(context.Background())
	// pgx.ConnConfig(pgx.ConnConfig{Config: pgx.ConnConfig{}})
	// pgxpool.NewWithConfig(context.Background(), &pgxpool.Config{ConnConfig: &pgx.ConnConfig{}})
	// api := API{}
}

// pgx.ConnConfig
