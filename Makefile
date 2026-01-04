APP_NAME := glpipeline

.PHONY: all build clean test

all: build

configure:
	@go install

build: configure
	@CGO_ENABLED=0 GOOS=linux go build -o $(APP_NAME) -v -ldflags="-s -w"

clean:
	@rm -f $(APP_NAME)

lint:
	@golangci-lint run

test: build
	@go test

install:
	@if [ $$(id -u) -eq 0 ] ; then \
		bindir="/usr/local/bin" ; \
		if [ ! -z "$${BINDIR}" ] ; then bindir="$${BINDIR}" ; fi ; \
		mkdir -p $${bindir} ; \
		cp $(APP_NAME) $${bindir}/ ; \
	else \
		mkdir -p $${HOME}/.local/bin ; \
		cp $(APP_NAME) $${HOME}/.local/bin/ ; \
	fi
