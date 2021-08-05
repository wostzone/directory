DIST_FOLDER=./dist
.DEFAULT_GOAL := help

FORCE: help

all: thingdir-pb ## Build package with binary distribution and config

test: FORCE ## Run tests (stop on first error, don't run parallel)
		go test -failfast -p 1 -v ./pkg/...

install: all  ## Install the plugin into ~/bin/wost/bin and config
	cp dist/bin/* ~/bin/wost/bin/
	cp -n dist/config/* ~/bin/wost/config/

clean: ## Clean distribution files
	go mod tidy
	go clean
	rm -f $(DIST_FOLDER)/certs/*
	rm -f $(DIST_FOLDER)/logs/*
	rm -f $(DIST_FOLDER)/bin/*
	rm -f $(DIST_FOLDER)/arm/*

thingdir-pb: ## Build thingdir-pb plugin
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/$@/main.go
	@echo "> SUCCESS. Plugin '$@' can be found at $(DIST_FOLDER)/bin/$@ and $(DIST_FOLDER)/arm/$@"

help: ## Show this help
		@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
