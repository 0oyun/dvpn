# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGEN=$(GOCMD) generate
GOGET=$(GOCMD) get
GOLIST=$(GOCMD) list
BINARY_NAME=miniDVPN

# version info
branch=$(shell git rev-parse --abbrev-ref HEAD)
commitID=$(shell git log --pretty=format:"%h" -1)
date=$(shell date +%Y%m%d)
importpath=github.com/toy-playground/miniDVPN/cmd
ldflags=-X ${importpath}.branch=${branch} -X ${importpath}.commitID=${commitID} -X ${importpath}.date=${date}

# export gomodule
export GO111MODULE=on

all: build

## build: build the binary with pre-packed static resource
build:
	@export GOPROXY=https://goproxy.cn,direct
	@go build -o $(BINARY_NAME) -trimpath -ldflags "${ldflags}"

