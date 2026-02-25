#!/usr/bin/env bash
set -euo pipefail

WHISPER_DIR="${WHISPER_DIR:-$HOME/whisper.cpp}"
MODEL_NAME="${MODEL_NAME:-ggml-large-v3-turbo.bin}"
MODEL_URL="https://huggingface.co/ggerganov/whisper.cpp/resolve/main/${MODEL_NAME}"

echo "=== Building whisper.cpp with CUDA support ==="

if [ ! -d "$WHISPER_DIR" ]; then
    git clone https://github.com/ggerganov/whisper.cpp.git "$WHISPER_DIR"
else
    echo "whisper.cpp already cloned at $WHISPER_DIR"
    cd "$WHISPER_DIR" && git pull
fi

cd "$WHISPER_DIR"

cmake -B build -DGGML_CUDA=ON
cmake --build build --config Release -j "$(nproc)"

echo "=== whisper-server built at $WHISPER_DIR/build/bin/whisper-server ==="

# Download model if not present
MODELS_DIR="$WHISPER_DIR/models"
mkdir -p "$MODELS_DIR"

if [ ! -f "$MODELS_DIR/$MODEL_NAME" ]; then
    echo "=== Downloading $MODEL_NAME ==="
    curl -L "$MODEL_URL" -o "$MODELS_DIR/$MODEL_NAME"
else
    echo "Model already exists at $MODELS_DIR/$MODEL_NAME"
fi

echo ""
echo "Done. Add to your config.toml:"
echo "  [whisper]"
echo "  server_path = \"$WHISPER_DIR/build/bin/whisper-server\""
echo "  model_path = \"$MODELS_DIR/$MODEL_NAME\""
