SHELL := /bin/sh

VERSION ?= 0.1.0
GO      ?= go
DIST    := dist
LDFLAGS := -s -w -X github.com/subsll/metacairn/internal/app.Version=$(VERSION)

.PHONY: all build local test race vet fmt linux checksums clean

all: build

## build: native binary for the current platform
build:
	@mkdir -p $(DIST)
	CGO_ENABLED=0 $(GO) build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" \
		-o $(DIST)/metacairn ./cmd/metacairn

## local: versioned native binary
local:
	@mkdir -p $(DIST)
	CGO_ENABLED=0 $(GO) build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" \
		-o $(DIST)/metacairn_$(VERSION)_$(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH) ./cmd/metacairn

## linux: static Linux binaries
linux:
	@mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" -o $(DIST)/metacairn_$(VERSION)_linux_amd64 ./cmd/metacairn
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" -o $(DIST)/metacairn_$(VERSION)_linux_arm64 ./cmd/metacairn
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 $(GO) build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" -o $(DIST)/metacairn_$(VERSION)_linux_armv7 ./cmd/metacairn
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GO) build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" -o $(DIST)/metacairn_$(VERSION)_linux_386 ./cmd/metacairn

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w .

checksums:
	@cd $(DIST) && \
	if command -v sha256sum >/dev/null 2>&1; then \
		sha256sum metacairn_* > SHA256SUMS; \
	else \
		shasum -a 256 metacairn_* > SHA256SUMS; \
	fi && cat SHA256SUMS

clean:
	rm -rf $(DIST)
