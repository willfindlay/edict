package audio

import "sync"

// RingBuffer is a thread-safe ring buffer for PCM int16 samples.
type RingBuffer struct {
	mu   sync.Mutex
	data []int16
	pos  int
	full bool
}

// NewRingBuffer creates a ring buffer that holds capacity samples.
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		data: make([]int16, capacity),
	}
}

// Write appends samples to the buffer, overwriting oldest data if full.
func (b *RingBuffer) Write(samples []int16) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, s := range samples {
		b.data[b.pos] = s
		b.pos = (b.pos + 1) % len(b.data)
		if b.pos == 0 && !b.full {
			b.full = true
		}
	}
}

// Len returns the number of samples currently in the buffer.
func (b *RingBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.full {
		return len(b.data)
	}
	return b.pos
}

// Drain returns all samples in chronological order and resets the buffer.
func (b *RingBuffer) Drain() []int16 {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := b.pos
	if b.full {
		n = len(b.data)
	}

	out := make([]int16, n)
	if b.full {
		// Copy from pos to end, then start to pos
		copied := copy(out, b.data[b.pos:])
		copy(out[copied:], b.data[:b.pos])
	} else {
		copy(out, b.data[:b.pos])
	}

	b.pos = 0
	b.full = false
	return out
}

// Recent returns the most recent n samples (or fewer if buffer has less).
func (b *RingBuffer) Recent(n int) []int16 {
	b.mu.Lock()
	defer b.mu.Unlock()

	total := b.pos
	if b.full {
		total = len(b.data)
	}
	if n > total {
		n = total
	}
	if n == 0 {
		return nil
	}

	out := make([]int16, n)
	if b.full {
		// Start reading from (pos - n) mod cap
		start := (b.pos - n + len(b.data)) % len(b.data)
		if start < b.pos {
			copy(out, b.data[start:b.pos])
		} else {
			copied := copy(out, b.data[start:])
			copy(out[copied:], b.data[:b.pos])
		}
	} else {
		copy(out, b.data[b.pos-n:b.pos])
	}
	return out
}

// Snapshot returns all samples in chronological order without resetting the buffer.
func (b *RingBuffer) Snapshot() []int16 {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := b.pos
	if b.full {
		n = len(b.data)
	}
	if n == 0 {
		return nil
	}

	out := make([]int16, n)
	if b.full {
		copied := copy(out, b.data[b.pos:])
		copy(out[copied:], b.data[:b.pos])
	} else {
		copy(out, b.data[:b.pos])
	}

	return out
}

// Reset clears the buffer without deallocating.
func (b *RingBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.pos = 0
	b.full = false
}
