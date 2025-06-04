# Beam Game Engine - GoReleaser Configuration

A GoReleaser configuration, handles compiling with Raylib for you.

## Setup 

### Windows & Linux Cross-Compilation

### 1. Install Dependencies

Run the setup script to install dependencies and tools for cross-compilation:

```bash
./scripts/setup-cross-compile.sh
```

### 2. Configure Your Project

Set environment variables to customize the build:

```bash
export PROJECT_NAME="your-game-name"        # Default: "beam-game"
export BINARY_NAME="your-binary-name"       # Default: PROJECT_NAME
export MAIN_PATH="./cmd/main.go"             # Default: "./cmd/main.go"
export GITHUB_OWNER="your-username"         # Default: "your-username"
export GITHUB_REPO="your-repo-name"         # Default: PROJECT_NAME
```

### 3. Build

Test builds for all platforms:
```bash
goreleaser build --snapshot --clean
```

Create a release:
```bash
goreleaser release --clean
```


### MacOS Compilation
To compile for macOS, you need to set the `GOOS` and `GOARCH` environment variables:

```bash
export GOOS=darwin
export GOARCH=arm64 # or "amd64" for Intel Macs
```

Set environment variables to customize the build:

```bash
export PROJECT_NAME="your-game-name"        # Default: "beam-game"
export BINARY_NAME="your-binary-name"       # Default: PROJECT_NAME
export MAIN_PATH="./cmd/main.go"             # Default: "./cmd/main.go"
export GITHUB_OWNER="your-username"         # Default: "your-username"
export GITHUB_REPO="your-repo-name"         # Default: PROJECT_NAME
```

Then run the build command:
```bash
goreleaser build --snapshot --clean
```