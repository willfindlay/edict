# Install edict build dependencies on Windows via MSYS2/MinGW.
# Run from an elevated PowerShell prompt.

$ErrorActionPreference = "Stop"

# Check for MSYS2
$msys2 = "C:\msys64\usr\bin\bash.exe"
if (-not (Test-Path $msys2)) {
    Write-Host "MSYS2 not found at $msys2"
    Write-Host "Install MSYS2 from https://www.msys2.org/ first."
    exit 1
}

Write-Host "Installing MinGW toolchain and build dependencies via pacman..."
& $msys2 -lc "pacman -S --noconfirm --needed mingw-w64-x86_64-gcc mingw-w64-x86_64-cmake mingw-w64-x86_64-pkg-config"

# Verify Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Go not found in PATH. Install Go from https://go.dev/dl/"
    exit 1
}

Write-Host ""
Write-Host "Dependencies installed. Make sure C:\msys64\mingw64\bin is in your PATH."
Write-Host "Build with: make build-windows (from WSL) or: go build -tags noaudio -o edict.exe ./cmd/edict"
