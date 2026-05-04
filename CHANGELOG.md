# Changelog

All notable changes to this library are documented here. The format
follows Keep a Changelog; the project follows Semantic Versioning.

## [Unreleased]

## [0.1.0] - 2026-05-04

Initial release.

### Added
- `pkg/audio` — Stage interface, Ring (SPSC), Levels (RMS).
- `pkg/denoise` — Denoiser interface, DeepFilter cgo wrapper, Passthrough,
  Stage adapter. Build tag `deepfilter` enables the cgo wrapper.
- `pkg/resample` — Decimate3 (48k → 16k), Interpolate3 (16k → 48k, NEW).
  Interpolate3 uses 69 FIR taps (empirically tuned for the
  decimate→interpolate roundtrip target of >25 dB SNR).
- `pkg/speaker` — Backend, Store, ComputeEmbedding, SpeakerGate (TSE),
  TSEExtractor interface, TSEStream OLA wrapper (NEW; sliding-window
  overlap-add for streaming use). Integration tests gated on the
  `speakerbeam` build tag and `TSE_MODEL_PATH` env var.
- `pkg/recorder` — Session, per-stage WAV writer.
- `pkg/sessions` — Manifest format. (Adapted from voice-keyboard with
  the dictation-specific `Transcripts` field dropped.)
- `third_party/deepfilter/` — vendored DeepFilterNet v0.5.6 binary
  (macOS arm64) + header + model archive.
- `scripts/fetch-models.sh` — pulls `tse_model.onnx` and
  `speaker_encoder.onnx` from a local voice-keyboard checkout. Future
  versions will pull from a published GitHub Release asset.
