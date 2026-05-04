// Adapted from voice-keyboard's core/internal/speaker/tse.go (MIT).
// Original: github.com/sihekuang/voice-keyboard
//
// Reorganised for public consumption: moved out of internal/, no
// behavioural changes.

package speaker

import (
	"context"

	"github.com/sihekuang/solo-audio-core/pkg/audio"
)

// TSEExtractor extracts the target speaker's audio from a mixed signal.
// Implementations capture the enrolled speaker reference at construction
// time so the per-call signature lines up with audio.Stage.
//
// Input:  mixed 16 kHz mono PCM (typically a chunker emission)
// Output: clean audio of the same length as mixed
//
// Implementations MUST also satisfy audio.Stage.
type TSEExtractor interface {
	audio.Stage
	Extract(ctx context.Context, mixed []float32) ([]float32, error)
}
