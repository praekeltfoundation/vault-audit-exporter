NAME := vault-audit-exporter
PROJECT := github.com/praekeltfoundation/$(NAME)
BIN_NAME := $(NAME)

NON_CMD_PACKAGES := $(shell go list ./... | fgrep -v '/cmd/')

VERSION := $(shell grep "const Version " version/version.go | sed -E 's/.*"(.+)"$$/\1/')
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_DIRTY := $(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)

LDFLAGS = "-X ${PROJECT}/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X ${PROJECT}/version.VersionPrerelease=${VSN_PRERELEASE}"
VSN_PRERELEASE = DEV

help:
	@echo "Management commands for ${NAME}:"
	@echo
	@echo 'Usage:'
	@echo '    make build           Compile the project.'
	@echo '    make clean           Clean the directory tree.'
	@echo '    make cover           Run tests and collect coverage data.'
	@echo '    make dep             Run dep ensure.'
	@echo '    make lint            Run gometalinter.'
	@echo '    make test            Run tests on a compiled project.'
	@echo

all: clean dep build test lint

default: build test

build:
	@echo "building ${BIN_NAME} ${VERSION} ${GIT_COMMIT}${GIT_DIRTY}"
	@echo "GOPATH=${GOPATH}"
	go build -ldflags ${LDFLAGS} -o bin/${BIN_NAME} cmd/${BIN_NAME}/main.go

dep:
	dep ensure -v

clean:
	@test ! -e bin/${BIN_NAME} || rm bin/${BIN_NAME}

test:
	go test -race ${NON_CMD_PACKAGES}

cover:
	go test -race -coverprofile=coverage.txt -covermode=atomic ${NON_CMD_PACKAGES}

lint:
	gometalinter --vendor --tests --deadline=120s ./...

.PHONY: all build clean cover default dep help lint test
