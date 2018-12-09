GO ?= go
GOFMT ?= gofmt "-s"
PACKAGES ?= $(shell $(GO) list ./... | grep -v /vendor/)
VETPACKAGES ?= $(shell $(GO) list ./... | grep -v /vendor/ | grep -v /examples/)
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")

all: install

install: deps
	govendor sync

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)	

vet:
	$(GO) vet $(VETPACKAGES)

deps:
	@hash govendor > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/kardianos/govendor; \
	fi

.PHONY: build
build:
	$(GO) build  $(GOFILES)

run:
	$(GO) run  $(GOFILES)
