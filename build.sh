#!/bin/bash

# Set the target OS and architecture for macOS
export GOOS=darwin
export GOARCH=arm64  # Change to "arm64" for Apple Silicon (M1/M2)

# Set the output binary name
OUTPUT="oltp_metrics_parser"

# Compile the Go program
echo "Compiling for macOS..."
go build -o $OUTPUT main.go

# Verify the build
if [ $? -eq 0 ]; then
    echo "Build successful: ./$OUTPUT"
else
    echo "Build failed."
    exit 1
fi
