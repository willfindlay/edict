.PHONY: build test test-v lint fmt clean deps

BINARY := edict.exe
PKG := ./...
# noaudio tag disables raylib's built-in miniaudio to avoid duplicate symbols with malgo
TAGS := -tags noaudio
WINDOWS_ENV := CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc

build:
	$(WINDOWS_ENV) go build $(TAGS) -o $(BINARY) ./cmd/edict

test:
	$(WINDOWS_ENV) go test $(TAGS) $(PKG) -count=1

test-v:
	$(WINDOWS_ENV) go test $(TAGS) $(PKG) -count=1 -v

lint:
	$(WINDOWS_ENV) golangci-lint run $(PKG)

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY)
	go clean -testcache

deps:
	./scripts/install-deps.sh
