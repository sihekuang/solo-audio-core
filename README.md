# solo-audio-core

Reusable Go audio building blocks for noise suppression and target speaker
extraction.

- **Denoise** — DeepFilterNet (Rust → C ABI, vendored).
- **TSE** — ConvTasNet Libri2Mix sep_noisy 16k + Wespeaker ECAPA-TDNN-512,
  loaded as a single combined ONNX model.
- **Streaming TSE** — sliding-window overlap-add wrapper for use in
  continuous-streaming pipelines.
- **Resample** — 48 kHz ↔ 16 kHz (FIR + decimate/interpolate by 3).
- **Ring buffers** — single-producer / single-consumer lock-free
  `[]float32` ring.
- **Per-stage WAV recording** — capture the output of every pipeline
  stage to disk for offline inspection.

## Installation

```bash
go get github.com/sihekuang/solo-audio-core@v0.1.0
```

For real denoise (not the passthrough fallback), build with:

```bash
go build -tags deepfilter ./...
```

The `deepfilter` tag enables the cgo wrapper around the vendored libdf.

## Models

Models are not shipped in the repo. Run `scripts/fetch-models.sh` to
download `tse_model.onnx` and `speaker_encoder.onnx`.

## Origin

Adapted from voice-keyboard's `internal/*` packages (MIT). See
`THIRD_PARTY_NOTICES.md`.

## Status

Pre-1.0. Public API may change between minor versions until v1.0.0.
