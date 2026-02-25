package audio

import (
	"sync"
	"testing"
)

func TestRingBufferWrite(t *testing.T) {
	b := NewRingBuffer(4)

	b.Write([]int16{1, 2, 3})
	if b.Len() != 3 {
		t.Errorf("expected len 3, got %d", b.Len())
	}

	got := b.Drain()
	want := []int16{1, 2, 3}
	assertSamples(t, got, want)
}

func TestRingBufferOverwrite(t *testing.T) {
	b := NewRingBuffer(4)

	b.Write([]int16{1, 2, 3, 4, 5, 6})
	if b.Len() != 4 {
		t.Errorf("expected len 4, got %d", b.Len())
	}

	got := b.Drain()
	want := []int16{3, 4, 5, 6}
	assertSamples(t, got, want)
}

func TestRingBufferWrapAround(t *testing.T) {
	b := NewRingBuffer(4)

	b.Write([]int16{1, 2, 3})
	b.Write([]int16{4, 5})
	// Buffer: [5, 2, 3, 4] with pos=1, full=true
	// Chronological: 2, 3, 4, 5

	got := b.Drain()
	want := []int16{2, 3, 4, 5}
	assertSamples(t, got, want)
}

func TestRingBufferRecent(t *testing.T) {
	b := NewRingBuffer(8)

	b.Write([]int16{10, 20, 30, 40, 50})

	got := b.Recent(3)
	want := []int16{30, 40, 50}
	assertSamples(t, got, want)
}

func TestRingBufferRecentOverflow(t *testing.T) {
	b := NewRingBuffer(4)
	b.Write([]int16{1, 2, 3, 4, 5})

	got := b.Recent(3)
	want := []int16{3, 4, 5}
	assertSamples(t, got, want)
}

func TestRingBufferRecentMoreThanAvailable(t *testing.T) {
	b := NewRingBuffer(8)
	b.Write([]int16{1, 2})

	got := b.Recent(10)
	want := []int16{1, 2}
	assertSamples(t, got, want)
}

func TestRingBufferRecentEmpty(t *testing.T) {
	b := NewRingBuffer(4)
	got := b.Recent(5)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestRingBufferDrainResets(t *testing.T) {
	b := NewRingBuffer(4)
	b.Write([]int16{1, 2, 3})
	b.Drain()

	if b.Len() != 0 {
		t.Errorf("expected len 0 after drain, got %d", b.Len())
	}

	b.Write([]int16{10, 20})
	got := b.Drain()
	want := []int16{10, 20}
	assertSamples(t, got, want)
}

func TestRingBufferSnapshot(t *testing.T) {
	b := NewRingBuffer(8)
	b.Write([]int16{10, 20, 30})

	got := b.Snapshot()
	want := []int16{10, 20, 30}
	assertSamples(t, got, want)

	// Snapshot should not reset the buffer
	if b.Len() != 3 {
		t.Errorf("expected len 3 after snapshot, got %d", b.Len())
	}

	// Drain should still return the same data
	got2 := b.Drain()
	assertSamples(t, got2, want)
}

func TestRingBufferSnapshotOverflow(t *testing.T) {
	b := NewRingBuffer(4)
	b.Write([]int16{1, 2, 3, 4, 5, 6})

	got := b.Snapshot()
	want := []int16{3, 4, 5, 6}
	assertSamples(t, got, want)

	// Buffer should still be full after snapshot
	if b.Len() != 4 {
		t.Errorf("expected len 4 after snapshot, got %d", b.Len())
	}
}

func TestRingBufferSnapshotEmpty(t *testing.T) {
	b := NewRingBuffer(4)
	got := b.Snapshot()
	if got != nil {
		t.Errorf("expected nil for empty snapshot, got %v", got)
	}
}

func TestRingBufferReset(t *testing.T) {
	b := NewRingBuffer(4)
	b.Write([]int16{1, 2, 3, 4})
	b.Reset()

	if b.Len() != 0 {
		t.Errorf("expected len 0, got %d", b.Len())
	}
}

func TestRingBufferConcurrency(t *testing.T) {
	b := NewRingBuffer(1024)
	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)
		go func(base int16) {
			defer wg.Done()
			for j := range 100 {
				b.Write([]int16{base + int16(j)})
			}
		}(int16(i * 1000))
	}

	wg.Wait()

	n := b.Len()
	if n != 1000 {
		t.Errorf("expected 1000 samples, got %d", n)
	}
}

func assertSamples(t *testing.T, got, want []int16) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("sample[%d]: got %d, want %d", i, got[i], want[i])
		}
	}
}
