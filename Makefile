SHELL := /bin/bash

GO=go

# Required Go Packages
# go install github.com/smallnest/gen@latest
GEN=gen
# go install golang.org/x/tools/cmd/stringer@latest

# go package to save db schema go bindings in
GO_PACKAGE=db
# db secretes
SECRETS=./.dbconn.env

.PHONY: dbgen, gogen, gen

dbgen:
	source $(SECRETS) && $(GEN) --sqltype=postgres --connstr="postgresql://$$USER:$$PASSWORD@$$SERVER" -d $$DB --table=users,refresh_tokens --model=$(GO_PACKAGE) --gorm --file_naming="gen_{{.}}" --no-json --overwrite

gogen:
	$(GO) generate ./...

gen: gogen
