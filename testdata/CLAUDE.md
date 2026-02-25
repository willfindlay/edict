# Test Project

Build with `make build` and test with `make test`.

## Architecture

Uses `goroutine` pools and `RingBuffer` for audio processing.
The `WhisperClient` sends requests to the transcription server.
Run `golangci-lint` for linting.

## Code Standards

- Go 1.25, modules enabled
- func ProcessAudio handles the pipeline
- type AudioConfig struct for settings
- var DefaultTimeout = 30s
