## Setup 

Works for Linux / Windows AMD64

1. Run the setup script to install dependencies and tools for cross-compilation
# ./scripts/setup-cross-compile.sh
2. goreleaser build --snapshot --clean
3. Run the test builds script to compile binaries for different platforms
# ./scripts/test-builds.sh
