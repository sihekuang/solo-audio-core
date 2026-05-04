package resample

import (
	"context"
	"math"
	"testing"
)

func TestInterpolate3_OutputRate(t *testing.T) {
	i := NewInterpolate3()
	if i.OutputRate() != 48000 {
		t.Fatalf("OutputRate = %d, want 48000", i.OutputRate())
	}
	if i.Name() != "upsample" {
		t.Fatalf("Name = %q, want %q", i.Name(), "upsample")
	}
}

func TestInterpolate3_LengthRatio(t *testing.T) {
	i := NewInterpolate3()
	in := make([]float32, 100)
	out, err := i.Process(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 300 {
		t.Fatalf("len(out) = %d, want 300", len(out))
	}
}

func TestInterpolate3_DCPassthrough(t *testing.T) {
	// A constant input should produce a near-constant output (modulo
	// FIR ringing) once the filter has settled.
	i := NewInterpolate3()
	in := make([]float32, 200)
	for k := range in {
		in[k] = 0.5
	}
	out, _ := i.Process(context.Background(), in)
	// After settle (~33 samples in zero-inserted output), the upsampled
	// stream should converge to ~0.5
	settled := out[200:]
	for k, v := range settled {
		if math.Abs(float64(v)-0.5) > 0.01 {
			t.Fatalf("out[%d] = %v, want ~0.5", 200+k, v)
		}
	}
}

func TestInterpolate3_DecimateRoundtrip(t *testing.T) {
	// 48k → decimate → 16k → interpolate → 48k. Output should
	// approximate input within filter tolerance once both filters
	// have settled.
	const n = 4800
	in := make([]float32, n)
	for k := range in {
		in[k] = float32(math.Sin(2 * math.Pi * 1000 * float64(k) / 48000))
	}
	d := NewDecimate3()
	mid, _ := d.Process(context.Background(), in)
	i := NewInterpolate3()
	out, _ := i.Process(context.Background(), mid)
	if len(out) != n {
		t.Fatalf("roundtrip len = %d, want %d", len(out), n)
	}
	var sumSq, errSq float64
	for k := 200; k < n-200; k++ {
		sumSq += float64(in[k]) * float64(in[k])
		d := float64(out[k] - in[k])
		errSq += d * d
	}
	snr := 10 * math.Log10(sumSq/errSq)
	if snr < 25 {
		t.Fatalf("roundtrip SNR = %.1f dB, want > 25 dB", snr)
	}
}
