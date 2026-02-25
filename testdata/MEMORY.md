# Project Memory

## Audio Pipeline

The `RingBuffer` uses a mutex for thread safety. Watch for `deadlock` issues
when `malgo` callbacks run on the audio thread.

## Whisper

Use `large-v3-turbo` model for best speed/accuracy. The `whisper-server`
binary needs CUDA support.

ContextScanner runs every 30 seconds to refresh the prompt.
