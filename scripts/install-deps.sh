#!/usr/bin/env bash
set -euo pipefail

echo "=== Installing edict build dependencies ==="

# Audio (PulseAudio/PipeWire)
sudo pacman -S --needed --noconfirm \
    pulseaudio \
    libpulse \
    alsa-lib

# X11 libs (needed by raylib and gohook via WSLg)
sudo pacman -S --needed --noconfirm \
    libx11 \
    libxcursor \
    libxrandr \
    libxinerama \
    libxi \
    libxkbcommon \
    libxcb \
    mesa \
    wayland

# Build tools
sudo pacman -S --needed --noconfirm \
    gcc \
    pkg-config \
    cmake \
    git

echo ""
echo "=== Dependencies installed ==="
