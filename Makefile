SHELL := /bin/bash

# go install github.com/smallnest/gen@latest
GEN=gen
GO_PACKAGE=db
SECRETS=./.dbconn.env

.PHONY: gen

gen:
	source $(SECRETS) && $(GEN) --sqltype=postgres --connstr="postgresql://$$USER:$$PASSWORD@$$SERVER" -d $$DB --table=users,refresh_tokens --model=$(GO_PACKAGE) --gorm --file_naming="gen_{{.}}" --no-json --overwrite
