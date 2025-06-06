#!/bin/bash
set -e

echo "Setting up Cross-Compilation Environment for Beam Game Engine"

# Update system packages
echo "Updating system packages..."
sudo apt-get update

# Install raylib dependencies for Linux
echo "Installing Raylib dependencies..."
sudo apt-get install -y \
    libgl1-mesa-dev \
    libxi-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libwayland-dev \
    libxkbcommon-dev \
    unzip \
    curl

# Install cross-compilation tools for Windows
echo "Installing Windows cross-compilation tools..."
sudo apt-get install -y \
    gcc-mingw-w64-x86-64 \
    gcc-mingw-w64 \
    gcc-multilib

# Download raylib.dll for Windows builds
echo "Setting up Raylib for Windows builds..."
cd /tmp
if [ -f "raylib-release.zip" ]; then
    rm raylib-release.zip
fi

echo "Downloading Raylib Windows binaries..."
curl -L -o raylib-release.zip \
    "https://github.com/raysan5/raylib/releases/download/5.0/raylib-5.0_win64_msvc16.zip"

echo "Extracting raylib.dll..."
if unzip -j raylib-release.zip "*/raylib.dll" -d "$HOME/" 2>/dev/null; then
    echo "Extracted raylib.dll to $HOME/"
elif unzip -j raylib-release.zip "raylib.dll" -d "$HOME/" 2>/dev/null; then
    echo "Extracted raylib.dll to $HOME/"
elif unzip -j raylib-release.zip "**/raylib.dll" -d "$HOME/" 2>/dev/null; then
    echo "Extracted raylib.dll to $HOME/"
else
    echo "Warning: Could not find raylib.dll in the archive"
    echo "Archive contents:"
    unzip -l raylib-release.zip
fi

rm raylib-release.zip

# Test cross-compilation setup
echo ""
echo "=========================================="
echo "Testing Cross-Compilation Setup"
echo "=========================================="
echo ""

echo "Available Windows cross-compilers:"
which x86_64-w64-mingw32-gcc || echo "x86_64-w64-mingw32-gcc not found"

echo ""
echo "Native Linux compiler:"
which gcc || echo "gcc not found"

echo ""
echo "Setup complete!"
echo ""
echo "Next steps:"
echo "1. Navigate to your project directory"
echo "2. Set environment variables (optional):"
echo "   export PROJECT_NAME=your-game-name"
echo "   export BINARY_NAME=your-binary-name"
echo "   export MAIN_PATH=./cmd/main.go"
echo "   export GITHUB_OWNER=your-username"
echo "   export GITHUB_REPO=your-repo-name"
echo "3. Run: goreleaser build --snapshot --clean"
echo ""
echo "Note: This setup supports Windows x64, Linux x64, and macOS ARM64 builds."
echo ""

echo "Cross-compilation environment setup complete!"