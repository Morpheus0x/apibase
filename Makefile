SHELL := /bin/bash

GO=go

# Required Go Packages
# go install github.com/smallnest/gen@latest
GEN=gen
# go install golang.org/x/tools/cmd/stringer@latest

# go package to save db schema go bindings in
GO_PACKAGE=db
# db secretes
DB_SECRETS=./.dbconn.env

# go install github.com/stephenafamo/bob/gen/bobgen-psql@latest
BOB=bobgen-psql
# db secrets
BOB_SECRETS=./.pgconn.yaml
BOB_SECRETS_STAGING=./.pgconn.staging.yaml

.PHONY: gormgen, dbgen, gogen, gen

gormgen:
	source $(DB_SECRETS) && $(GEN) --sqltype=postgres --connstr="postgresql://$$USER:$$PASSWORD@$$SERVER" -d $$DB --table=users,refresh_tokens --model=$(GO_PACKAGE) --gorm --file_naming="gen_{{.}}" --no-json --overwrite

dbgen:
	$(BOB) -c $(BOB_SECRETS)

dbgenstaging:
	$(BOB) -c $(BOB_SECRETS_STAGING)

gogen:
	$(GO) generate ./...

gen: gogen
