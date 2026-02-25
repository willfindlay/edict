# Edict

Speech-to-text for Claude Code. Captures microphone audio, transcribes via local GPU-accelerated Whisper, and types the result into the focused terminal. Claude Code-aware: scans running processes, CLAUDE.md files, memory files, and skill definitions to build domain-specific Whisper prompts.

## Build & Test

```
make build       # Build binary to ./edict
make test        # Run all tests
make lint        # Run golangci-lint
make fmt         # Format with gofumpt
```

## Architecture

Goroutines: main (raylib overlay), hotkey (gohook listener), pipeline (capture -> transcribe -> type), context (periodic Claude Code scan).

## Code Standards

- Go 1.25, modules enabled
- `golangci-lint` for linting
- `gofumpt` for formatting
- Tests use `testing` stdlib, no test frameworks
- CGO required (malgo, raylib, gohook)
