# Changelog

All notable changes to this library are documented here. The format
follows Keep a Changelog; the project follows Semantic Versioning.

## [Unreleased]

## [0.1.0] - 2026-05-04 (planned)

Initial release.

### Added
- `pkg/audio` — Stage interface, Ring (SPSC), Levels (RMS).
- `pkg/denoise` — Denoiser interface, DeepFilter cgo wrapper, Passthrough,
  Stage adapter.
- `pkg/resample` — Decimate3 (48k → 16k), Interpolate3 (16k → 48k, NEW).
- `pkg/speaker` — Backend, Store, ComputeEmbedding, SpeakerGate (TSE),
  TSEExtractor interface, TSEStream OLA wrapper (NEW).
- `pkg/recorder` — Session, per-stage WAV writer.
- `pkg/sessions` — Manifest format.
