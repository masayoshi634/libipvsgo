NAME     := libipvs
VERSION  := v0.1
REVISION := $(shell git rev-parse --short HEAD)

SRCS    := $(shell find . -type f -name '*.go')
LDFLAGS := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -extldflags \"-static\""

CREDITS = vendor/CREDITS

bin/$(NAME): $(SRCS) deps
	go build -a -tags netgo -installsuffix netgo $(LDFLAGS)

.PHONY: go-dep
go-dep:
ifeq ($(shell command -v dep 2> /dev/null),)
	go get -u github.com/golang/dep/cmd/dep
endif

.PHONY: deps
deps: go-dep
	dep ensure

.PHONY: clean
clean:
	rm -rf vendor/*

.PHONY: credits
credits: deps
	scripts/credits > $(CREDITS)

.PHONY: test
test: deps
	sudo -E go test
