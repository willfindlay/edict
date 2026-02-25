# Edict

Speech-to-text for Claude Code. Captures microphone audio, transcribes via local GPU-accelerated Whisper, and types the result into the focused terminal. Claude Code-aware: scans running processes, CLAUDE.md files, memory files, and skill definitions to build domain-specific Whisper prompts.

## Build & Test

```
make build       # Build binary to ./edict
make test        # Run all tests
make test-v      # Run all tests (verbose)
make lint        # Run golangci-lint
make fmt         # Format with gofmt
make deps        # Install system dependencies
make whisper     # Build whisper.cpp with CUDA + download model
```

## Architecture

Goroutines: main (raylib overlay), hotkey (gohook listener), pipeline (capture -> transcribe -> type), context (periodic Claude Code scan at 30s intervals).

## Build Notes

- `noaudio` build tag is required (and set in Makefile) to avoid duplicate miniaudio symbols between malgo and raylib-go
- CGO required: malgo (miniaudio), raylib-go (OpenGL), gohook (libuiohook)
- Platform: Linux/WSL2 (WSLg for audio + display)

## Code Standards

- Go 1.25, modules enabled
- `golangci-lint` for linting
- `gofmt` for formatting
- Tests use `testing` stdlib, no test frameworks
