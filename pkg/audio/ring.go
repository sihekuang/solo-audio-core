package audio

import (
	"sync/atomic"
)

// Ring is a single-producer / single-consumer lock-free ring buffer of
// float32 samples. One goroutine may call Push; one (different)
// goroutine may call Pull; concurrent calls into the same method by
// multiple goroutines are NOT supported. Capacity is rounded up to the
// next power of two so the modulo is a bitmask. The ring keeps
// capacity-1 usable slots to disambiguate full from empty.
//
// Safe to use as the bridge between an audio thread (producer) and a
// pump thread (consumer).
type Ring struct {
	buf  []float32
	mask uint64 // capacity - 1, used for cheap modulo
	head atomic.Uint64 // next write index (producer only writes)
	tail atomic.Uint64 // next read index  (consumer only writes)
}

// NewRing returns a Ring with capacity rounded up to the next power of
// two (minimum 1). The maximum number of samples that can be pushed
// without an interleaved pull is capacity-1.
func NewRing(capacity int) *Ring {
	if capacity < 1 {
		capacity = 1
	}
	cap2 := 1
	for cap2 < capacity {
		cap2 <<= 1
	}
	return &Ring{
		buf:  make([]float32, cap2),
		mask: uint64(cap2 - 1),
	}
}

// Capacity returns the rounded-up capacity (power of two).
func (r *Ring) Capacity() int { return len(r.buf) }

// Push writes up to len(in) samples into the ring and returns the
// number actually written. Returns less than len(in) when the ring is
// near full. Lock-free; safe to call from one producer goroutine.
func (r *Ring) Push(in []float32) int {
	if len(in) == 0 {
		return 0
	}
	tail := r.tail.Load()
	head := r.head.Load()
	free := int(uint64(len(r.buf)) - 1 - (head - tail))
	n := len(in)
	if n > free {
		n = free
	}
	for i := 0; i < n; i++ {
		r.buf[(head+uint64(i))&r.mask] = in[i]
	}
	r.head.Store(head + uint64(n))
	return n
}

// Pull reads up to len(out) samples from the ring and returns the
// number actually written into out. Returns less than len(out) when
// the ring is near empty. Lock-free; safe to call from one consumer
// goroutine.
func (r *Ring) Pull(out []float32) int {
	if len(out) == 0 {
		return 0
	}
	head := r.head.Load()
	tail := r.tail.Load()
	avail := int(head - tail)
	n := len(out)
	if n > avail {
		n = avail
	}
	for i := 0; i < n; i++ {
		out[i] = r.buf[(tail+uint64(i))&r.mask]
	}
	r.tail.Store(tail + uint64(n))
	return n
}

// Available returns the number of samples currently in the ring.
// Snapshot value — may be stale by the time the caller reads it.
func (r *Ring) Available() int {
	return int(r.head.Load() - r.tail.Load())
}
