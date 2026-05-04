package audio

import (
	"sync"
	"testing"
)

func TestRing_PushPull_RoundTrip(t *testing.T) {
	r := NewRing(16)
	in := []float32{1, 2, 3, 4, 5}
	if n := r.Push(in); n != 5 {
		t.Fatalf("Push returned %d, want 5", n)
	}
	out := make([]float32, 5)
	if n := r.Pull(out); n != 5 {
		t.Fatalf("Pull returned %d, want 5", n)
	}
	for i := range in {
		if out[i] != in[i] {
			t.Errorf("out[%d]=%v want %v", i, out[i], in[i])
		}
	}
}

func TestRing_Push_Saturates(t *testing.T) {
	r := NewRing(4) // capacity rounds to 4
	in := make([]float32, 10)
	for i := range in {
		in[i] = float32(i)
	}
	n := r.Push(in)
	// SPSC ring keeps capacity-1 usable to disambiguate full/empty.
	if n != 3 {
		t.Fatalf("Push returned %d, want 3 (capacity-1)", n)
	}
}

func TestRing_Pull_FromEmpty(t *testing.T) {
	r := NewRing(4)
	out := make([]float32, 3)
	if n := r.Pull(out); n != 0 {
		t.Fatalf("Pull from empty returned %d, want 0", n)
	}
}

func TestRing_PowerOfTwoCapacity(t *testing.T) {
	cases := []struct{ in, want int }{
		{1, 1}, {2, 2}, {3, 4}, {5, 8}, {1024, 1024}, {1025, 2048},
	}
	for _, c := range cases {
		r := NewRing(c.in)
		if got := r.Capacity(); got != c.want {
			t.Errorf("NewRing(%d).Capacity() = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestRing_TwoGoroutineSoak(t *testing.T) {
	const total = 100_000
	r := NewRing(1024)
	var wg sync.WaitGroup
	wg.Add(2)

	produced := make([]float32, total)
	for i := range produced {
		produced[i] = float32(i)
	}

	consumed := make([]float32, 0, total)

	go func() {
		defer wg.Done()
		i := 0
		for i < total {
			n := r.Push(produced[i:])
			i += n
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]float32, 64)
		for len(consumed) < total {
			n := r.Pull(buf)
			consumed = append(consumed, buf[:n]...)
		}
	}()

	wg.Wait()
	if len(consumed) != total {
		t.Fatalf("consumed %d, want %d", len(consumed), total)
	}
	for i := 0; i < total; i++ {
		if consumed[i] != float32(i) {
			t.Fatalf("consumed[%d] = %v, want %v (corruption)", i, consumed[i], float32(i))
		}
	}
}
