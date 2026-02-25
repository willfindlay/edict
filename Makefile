.PHONY: build build-windows test lint fmt clean install run deps whisper test-v

BINARY := edict
PKG := ./...
# noaudio tag disables raylib's built-in miniaudio to avoid duplicate symbols with malgo
TAGS := -tags noaudio

build:
	CGO_ENABLED=1 go build $(TAGS) -o $(BINARY) ./cmd/edict

build-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build $(TAGS) -o $(BINARY).exe ./cmd/edict

test:
	CGO_ENABLED=1 go test $(TAGS) $(PKG) -count=1

test-v:
	CGO_ENABLED=1 go test $(TAGS) $(PKG) -count=1 -v

lint:
	golangci-lint run $(TAGS) $(PKG)

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY) $(BINARY).exe
	go clean -testcache

install: build
	install -m 755 $(BINARY) $(GOPATH)/bin/$(BINARY)

run: build
	./$(BINARY)

deps:
	./scripts/install-deps.sh

whisper:
	./scripts/install-whisper.sh
