# Edict

Speech-to-text for Claude Code. Captures microphone audio, transcribes via local GPU-accelerated Whisper, and types the result into the focused terminal. Uses a static Whisper prompt tuned for developer speech, slash commands, and shell paths.

## Build & Test

```
make build           # Cross-compile to ./edict.exe (requires MinGW)
make test            # Run all tests (cross-compiled, runs via binfmt_misc)
make test-v          # Run all tests (verbose)
make lint            # Run golangci-lint
make fmt             # Format with gofmt
make deps            # Install build dependencies (WSL2)
docker compose up -d # Start GPU-accelerated whisper-server
```

## Architecture

Goroutines: main (raylib overlay), hotkey (gohook listener), pipeline (capture -> transcribe -> type).

## Platform Support

Edict runs as a Windows .exe, cross-compiled from WSL2. Platform-specific code uses `_windows.go` build tags in:
- `internal/output/` - SendInput typing backend
- `internal/hotkey/` - Windows VK codes
- `internal/config/` - Windows defaults and backend validation
- `cmd/edict/` - signal handling, config paths

## Build Notes

- `noaudio` build tag is required (and set in Makefile) to avoid duplicate miniaudio symbols between malgo and raylib-go
- All Makefile targets use `GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc`
- CGO required: malgo (miniaudio), raylib-go (OpenGL), gohook (libuiohook)
- WSL2 host needs MinGW cross-compiler, audio libs, and X11 libs for CGO linking
- Tests cross-compile to `.exe` and run via WSL2 binfmt_misc

## Code Standards

- Go 1.25, modules enabled
- `golangci-lint` for linting
- `gofmt` for formatting
- Tests use `testing` stdlib, no test frameworks
