package audio

import "sync"

// SampleBuffer is a thread-safe growable buffer for PCM int16 samples.
// Unlike a ring buffer, it never discards data, supporting unlimited
// recording duration.
type SampleBuffer struct {
	mu      sync.Mutex
	samples []int16
}

// NewSampleBuffer creates a growable sample buffer pre-allocated for
// approximately 30 seconds of 16kHz mono audio.
func NewSampleBuffer() *SampleBuffer {
	return &SampleBuffer{
		samples: make([]int16, 0, 16000*30),
	}
}

// Write appends samples to the buffer.
func (b *SampleBuffer) Write(samples []int16) {
	b.mu.Lock()
	b.samples = append(b.samples, samples...)
	b.mu.Unlock()
}

// Len returns the number of samples currently in the buffer.
func (b *SampleBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.samples)
}

// Drain returns all samples and resets the buffer. The returned slice
// is owned by the caller; the buffer allocates a fresh backing array.
func (b *SampleBuffer) Drain() []int16 {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := b.samples
	b.samples = make([]int16, 0, cap(out))
	return out
}

// Recent returns the most recent n samples (or fewer if the buffer has less).
func (b *SampleBuffer) Recent(n int) []int16 {
	b.mu.Lock()
	defer b.mu.Unlock()

	if n > len(b.samples) {
		n = len(b.samples)
	}
	if n == 0 {
		return nil
	}

	out := make([]int16, n)
	copy(out, b.samples[len(b.samples)-n:])
	return out
}

// Snapshot returns a copy of all samples without resetting the buffer.
func (b *SampleBuffer) Snapshot() []int16 {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.samples) == 0 {
		return nil
	}

	out := make([]int16, len(b.samples))
	copy(out, b.samples)
	return out
}

// Reset clears the buffer without deallocating.
func (b *SampleBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.samples = b.samples[:0]
}
