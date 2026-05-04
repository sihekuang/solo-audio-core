// Adapted from voice-keyboard's core/internal/denoise/passthrough.go (MIT).
// Original: github.com/sihekuang/voice-keyboard
//
// Reorganised for public consumption: moved out of internal/, no
// behavioural changes.

package denoise

// Passthrough is a no-op Denoiser used when the user disables noise
// suppression. It returns the input frame unchanged.
type Passthrough struct{}

func NewPassthrough() *Passthrough { return &Passthrough{} }

func (p *Passthrough) Process(frame []float32) []float32 {
	out := make([]float32, len(frame))
	copy(out, frame)
	return out
}

func (p *Passthrough) Close() error { return nil }
