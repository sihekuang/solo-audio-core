package speaker

import (
	"context"
	"math"

	"github.com/sihekuang/solo-audio-core/pkg/audio"
)

// TSEStreamOptions configures a TSEStream. SampleRate must match the
// inner extractor's expected input rate (16000 for SpeakerGate).
// WindowMs and HopMs are in milliseconds; WindowMs > HopMs > 0.
type TSEStreamOptions struct {
	SampleRate int
	WindowMs   int
	HopMs      int
}

// TSEStream wraps any audio.Stage (typically *SpeakerGate) with a
// sliding-window overlap-add buffer so non-causal extractors can be
// used in a continuous stream. The inner stage is called once per
// window; outputs are crossfaded with the previous window using a
// Hann window so window boundaries don't audibly click.
//
// Steady-state added latency = WindowMs - HopMs.
type TSEStream struct {
	inner      audio.Stage
	sampleRate int
	windowSize int
	hopSize    int

	inBuf   []float32 // accumulated input
	overlap []float32 // last (windowSize - hopSize) samples of previous output
	hann    []float32 // Hann window over windowSize samples (precomputed)

	hasFirst bool
}

// NewTSEStream returns a streaming wrapper. Options.WindowMs and HopMs
// are clamped to >= 1.
func NewTSEStream(inner audio.Stage, opts TSEStreamOptions) *TSEStream {
	if opts.SampleRate <= 0 {
		opts.SampleRate = 16000
	}
	if opts.WindowMs <= 0 {
		opts.WindowMs = 1000
	}
	if opts.HopMs <= 0 {
		opts.HopMs = 250
	}
	windowSize := opts.WindowMs * opts.SampleRate / 1000
	hopSize := opts.HopMs * opts.SampleRate / 1000
	if hopSize > windowSize {
		hopSize = windowSize
	}
	return &TSEStream{
		inner:      inner,
		sampleRate: opts.SampleRate,
		windowSize: windowSize,
		hopSize:    hopSize,
		hann:       hannWindow(windowSize),
	}
}

func hannWindow(n int) []float32 {
	w := make([]float32, n)
	if n == 1 {
		w[0] = 1
		return w
	}
	for i := 0; i < n; i++ {
		w[i] = float32(0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1))))
	}
	return w
}

// Name implements audio.Stage. The inner extractor's name is hidden so
// the recorder/tap names stay stable as "tse" regardless of backend.
func (s *TSEStream) Name() string { return "tse" }

// OutputRate implements audio.Stage. Same as the inner extractor.
func (s *TSEStream) OutputRate() int { return s.sampleRate }

// LastSimilarity returns the inner extractor's last similarity, or 1.0
// when the inner extractor doesn't expose one.
func (s *TSEStream) LastSimilarity() float32 {
	if g, ok := s.inner.(interface{ LastSimilarity() float32 }); ok {
		return g.LastSimilarity()
	}
	return 1.0
}

// Process accumulates input samples into a buffer, runs the inner
// extractor whenever a full window is ready, and emits hopSize
// crossfaded samples per inner call. May return zero samples while
// the first window fills.
func (s *TSEStream) Process(ctx context.Context, in []float32) ([]float32, error) {
	if len(in) > 0 {
		s.inBuf = append(s.inBuf, in...)
	}
	var out []float32
	for len(s.inBuf) >= s.windowSize {
		win := make([]float32, s.windowSize)
		copy(win, s.inBuf[:s.windowSize])
		extracted, err := s.inner.Process(ctx, win)
		if err != nil {
			return nil, err
		}
		emit := make([]float32, s.hopSize)
		olen := s.windowSize - s.hopSize
		if s.hasFirst && olen > 0 {
			// OLA: blend extracted[olen:] with overlap[:hopSize]
			// using the Hann rising/falling halves around the boundary.
			for i := 0; i < s.hopSize; i++ {
				w := s.hann[olen+i]
				wPrev := s.hann[olen-1-i]
				sum := extracted[olen+i]*w + s.overlap[i]*wPrev
				if w+wPrev > 0 {
					sum /= (w + wPrev)
				}
				emit[i] = sum
			}
		} else {
			copy(emit, extracted[:s.hopSize])
			s.hasFirst = true
		}
		out = append(out, emit...)
		// New overlap = trailing (windowSize - hopSize) samples of extracted.
		if olen > 0 {
			newOverlap := make([]float32, olen)
			copy(newOverlap, extracted[s.hopSize:])
			s.overlap = newOverlap
		}
		// Slide inBuf forward by hopSize.
		copy(s.inBuf, s.inBuf[s.hopSize:])
		s.inBuf = s.inBuf[:len(s.inBuf)-s.hopSize]
	}
	return out, nil
}

// Flush zero-pads any residual to one final window, runs the extractor,
// and emits the trailing samples.
func (s *TSEStream) Flush(ctx context.Context) ([]float32, error) {
	if len(s.inBuf) == 0 && !s.hasFirst {
		return nil, nil
	}
	pad := make([]float32, s.windowSize)
	copy(pad, s.inBuf)
	s.inBuf = s.inBuf[:0]
	extracted, err := s.inner.Process(ctx, pad)
	if err != nil {
		return nil, err
	}
	return extracted, nil
}

var (
	_ audio.Stage   = (*TSEStream)(nil)
	_ audio.Flusher = (*TSEStream)(nil)
)
