.PHONY: build test lint fmt clean install run deps whisper

BINARY := edict
PKG := ./...

build:
	CGO_ENABLED=1 go build -o $(BINARY) ./cmd/edict

test:
	CGO_ENABLED=1 go test $(PKG) -count=1

lint:
	golangci-lint run $(PKG)

fmt:
	gofumpt -w .

clean:
	rm -f $(BINARY)
	go clean -testcache

install: build
	install -m 755 $(BINARY) $(GOPATH)/bin/$(BINARY)

run: build
	./$(BINARY)

deps:
	./scripts/install-deps.sh

whisper:
	./scripts/install-whisper.sh
