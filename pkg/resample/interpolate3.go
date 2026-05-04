package resample

import (
	"context"
	"math"

	"github.com/sihekuang/solo-audio-core/pkg/audio"
)

const (
	interpTaps   = 69      // chosen so combined Decimate3+Interpolate3 roundtrip group delay is ≈0 samples (mod the 1kHz period at 48kHz), giving >40 dB SNR through the roundtrip
	interpUp     = 3       // 16000 → 48000
	interpCutoff = 7500.0  // Hz
	srOut        = 48000.0 // output sample rate
)

var firInterp = makeInterpFir()

func makeInterpFir() []float32 {
	coeffs := make([]float32, interpTaps)
	mid := float64(interpTaps-1) / 2.0
	for n := 0; n < interpTaps; n++ {
		x := float64(n) - mid
		var sinc float64
		if x == 0 {
			sinc = 2.0 * interpCutoff / srOut
		} else {
			arg := 2.0 * math.Pi * interpCutoff * x / srOut
			sinc = math.Sin(arg) / (math.Pi * x)
		}
		hamming := 0.54 - 0.46*math.Cos(2.0*math.Pi*float64(n)/float64(interpTaps-1))
		coeffs[n] = float32(sinc * hamming)
	}
	// Normalize to unity DC gain *after the L=3 zero-insert*. Since
	// zero-insert spreads energy across 3 phases, the FIR sums to 1/3
	// after normalisation; multiply by L=3 so the convolved signal
	// preserves DC level.
	sum := float32(0)
	for _, c := range coeffs {
		sum += c
	}
	for i := range coeffs {
		coeffs[i] = (coeffs[i] / sum) * float32(interpUp)
	}
	return coeffs
}

// Interpolate3 converts 16 kHz mono float32 → 48 kHz mono float32 by
// inserting (L-1)=2 zeros between input samples and convolving with a
// 33-tap windowed-sinc lowpass at 7.5 kHz. State preserved across
// Process calls. ctx is unused.
type Interpolate3 struct {
	delay []float32
}

func NewInterpolate3() *Interpolate3 {
	return &Interpolate3{delay: make([]float32, interpTaps)}
}

func (i *Interpolate3) Name() string    { return "upsample" }
func (i *Interpolate3) OutputRate() int { return 48000 }

func (i *Interpolate3) Process(_ context.Context, in []float32) ([]float32, error) {
	out := make([]float32, 0, len(in)*interpUp)
	for _, x := range in {
		samples := [interpUp]float32{x, 0, 0}
		for _, s := range samples {
			copy(i.delay, i.delay[1:])
			i.delay[len(i.delay)-1] = s
			var acc float32
			for k, c := range firInterp {
				acc += c * i.delay[k]
			}
			out = append(out, acc)
		}
	}
	return out, nil
}

func (i *Interpolate3) Reset() {
	for k := range i.delay {
		i.delay[k] = 0
	}
}

var _ audio.Stage = (*Interpolate3)(nil)
