package speaker

import (
	"context"
	"math"
	"testing"

	"github.com/sihekuang/solo-audio-core/pkg/audio"
)

// fakeExtractor returns the input scaled by 0.5 (so we can verify the
// window pipeline correctly threads samples through).
type fakeExtractor struct {
	calls int
}

func (f *fakeExtractor) Name() string    { return "fake_tse" }
func (f *fakeExtractor) OutputRate() int { return 16000 }
func (f *fakeExtractor) Process(_ context.Context, in []float32) ([]float32, error) {
	f.calls++
	out := make([]float32, len(in))
	for i, v := range in {
		out[i] = v * 0.5
	}
	return out, nil
}

func TestTSEStream_Identity(t *testing.T) {
	// With windowSize == hopSize and the fake's identity-half scaling,
	// output should equal input * 0.5 once steady-state.
	fake := &fakeExtractor{}
	s := NewTSEStream(fake, TSEStreamOptions{
		SampleRate: 16000,
		WindowMs:   100, // 1600 samples
		HopMs:      100, // 1600 samples — no overlap
	})
	in := make([]float32, 16000) // 1 second
	for i := range in {
		in[i] = 1.0
	}
	out, err := s.Process(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 16000 {
		t.Fatalf("len(out) = %d, want 16000", len(out))
	}
	for i, v := range out {
		if math.Abs(float64(v)-0.5) > 1e-6 {
			t.Fatalf("out[%d] = %v, want 0.5", i, v)
		}
	}
}

func TestTSEStream_FillsThenEmits(t *testing.T) {
	// With windowSize > hopSize, the stage should buffer until the first
	// window is full before emitting anything.
	fake := &fakeExtractor{}
	s := NewTSEStream(fake, TSEStreamOptions{
		SampleRate: 16000,
		WindowMs:   1000, // 16000 samples
		HopMs:      250,  // 4000 samples
	})
	in := make([]float32, 8000)
	out, _ := s.Process(context.Background(), in)
	if len(out) != 0 {
		t.Fatalf("len(out) before window full = %d, want 0", len(out))
	}
	out, _ = s.Process(context.Background(), in)
	if len(out) != 4000 {
		t.Fatalf("first emission = %d samples, want 4000 (one hop)", len(out))
	}
	if fake.calls != 1 {
		t.Fatalf("fake called %d times, want 1", fake.calls)
	}
}

func TestTSEStream_Flush(t *testing.T) {
	fake := &fakeExtractor{}
	s := NewTSEStream(fake, TSEStreamOptions{
		SampleRate: 16000,
		WindowMs:   1000,
		HopMs:      250,
	})
	in := make([]float32, 5000)
	for i := range in {
		in[i] = 1.0
	}
	_, _ = s.Process(context.Background(), in)
	out, err := s.Flush(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("Flush emitted nothing — residual was lost")
	}
	if fake.calls < 1 {
		t.Fatal("Flush did not run the extractor on residual")
	}
}

func TestTSEStream_NameAndRate(t *testing.T) {
	fake := &fakeExtractor{}
	s := NewTSEStream(fake, TSEStreamOptions{SampleRate: 16000, WindowMs: 1000, HopMs: 250})
	if s.Name() != "tse" {
		t.Fatalf("Name = %q, want %q", s.Name(), "tse")
	}
	if s.OutputRate() != 16000 {
		t.Fatalf("OutputRate = %d, want 16000", s.OutputRate())
	}
}

func TestTSEStream_ImplementsStage(t *testing.T) {
	var _ audio.Stage = NewTSEStream(&fakeExtractor{}, TSEStreamOptions{SampleRate: 16000, WindowMs: 1000, HopMs: 250})
	var _ audio.Flusher = NewTSEStream(&fakeExtractor{}, TSEStreamOptions{SampleRate: 16000, WindowMs: 1000, HopMs: 250})
}
