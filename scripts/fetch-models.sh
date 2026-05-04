#!/usr/bin/env bash
# scripts/fetch-models.sh
#
# Downloads ONNX models needed for the speakerbeam-tagged tests and
# any consumer using pkg/speaker.SpeakerGate. Models are cached under
# `models/` (gitignored).
set -euo pipefail

MODELS_DIR="$(cd "$(dirname "$0")/.." && pwd)/models"
mkdir -p "$MODELS_DIR"

# v0.1.0: source from a local voice-keyboard checkout if present.
# Future versions: download from a GitHub Release asset.
VK_BUILD_MODELS="${VK_BUILD_MODELS:-$HOME/Documents/Projects/voice-keyboard/core/build/models}"

for f in tse_model.onnx speaker_encoder.onnx; do
  if [ -f "$MODELS_DIR/$f" ]; then
    echo "[ok] $f already present"
    continue
  fi
  if [ -f "$VK_BUILD_MODELS/$f" ]; then
    cp "$VK_BUILD_MODELS/$f" "$MODELS_DIR/$f"
    echo "[copied] $f from $VK_BUILD_MODELS"
  else
    echo "[error] $f not found at $VK_BUILD_MODELS — set VK_BUILD_MODELS or" >&2
    echo "        download from a future GitHub Release of solo-audio-core." >&2
    exit 1
  fi
done

echo "Models ready in $MODELS_DIR"
