#!/usr/bin/env bash
set -euo pipefail

echo "=== Installing edict system dependencies ==="

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

# Input simulation
sudo pacman -S --needed --noconfirm \
    ydotool \
    xdotool

# Build tools
sudo pacman -S --needed --noconfirm \
    gcc \
    pkg-config \
    cmake \
    git

echo ""
echo "=== Dependencies installed ==="
echo "You may need to start ydotoold: sudo systemctl enable --now ydotool"
