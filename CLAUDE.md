# Edict

Speech-to-text for Claude Code. Captures microphone audio, transcribes via local GPU-accelerated Whisper, and types the result into the focused terminal. Claude Code-aware: scans running processes, CLAUDE.md files, memory files, and skill definitions to build domain-specific Whisper prompts.

## Build & Test

```
make build           # Build binary to ./edict (Linux)
make build-windows   # Cross-compile to ./edict.exe (requires MinGW)
make test            # Run all tests
make test-v          # Run all tests (verbose)
make lint            # Run golangci-lint
make fmt             # Format with gofmt
make deps            # Install system dependencies (Linux)
make whisper         # Build whisper.cpp with CUDA + download model
```

## Architecture

Goroutines: main (raylib overlay), hotkey (gohook listener), pipeline (capture -> transcribe -> type), context (periodic Claude Code scan at 30s intervals).

## Platform Support

Linux and Windows are supported via `_linux.go` / `_windows.go` build-tagged files. Platform-specific code is isolated in:
- `internal/output/` - ydotool/xdotool (Linux) vs SendInput (Windows)
- `internal/hotkey/` - Linux scancodes vs Windows VK codes
- `internal/config/` - defaults and backend validation per platform
- `internal/context/` - /proc scanning (Linux) vs WSL shell-in (Windows)
- `cmd/edict/` - signal handling, config paths, home dir resolution

On Windows, edict shells into WSL to detect Claude Code processes and reads context files via `\\wsl.localhost\<distro>\...` UNC paths. The `[context]` config section controls WSL distro and home path.

## Build Notes

- `noaudio` build tag is required (and set in Makefile) to avoid duplicate miniaudio symbols between malgo and raylib-go
- CGO required: malgo (miniaudio), raylib-go (OpenGL), gohook (libuiohook)
- Linux: WSLg for audio + display, or native X11/PipeWire
- Windows: MSYS2/MinGW toolchain for CGO cross-compilation, WASAPI for audio

## Code Standards

- Go 1.25, modules enabled
- `golangci-lint` for linting
- `gofmt` for formatting
- Tests use `testing` stdlib, no test frameworks
