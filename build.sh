#!/bin/bash

# Set the target OS and architecture
export GOOS=darwin
export GOARCH=arm64  # Change to "amd64" for Intel Macs

# Define output directory
OUTPUT_DIR="./app"

# Ensure output directory exists
mkdir -p "$OUTPUT_DIR"

# List of programs to build (directories containing main.go)
PROGRAMS=("extract_metrics" "metrics_loadgen")
PROGRAM_PATHS=("src/extract_metrics" "src/metrics_loadgen")

# Compile each program
for i in "${!PROGRAMS[@]}"; do
    PROGRAM="${PROGRAMS[$i]}"
    PROGRAM_DIR="${PROGRAM_PATHS[$i]}"
    
    echo "🔧 Changing directory to ${PROGRAM_DIR}"
    cd "$PROGRAM_DIR" || { echo "❌ Failed to change directory to $PROGRAM_DIR"; exit 1; }

    echo "📦 Building $PROGRAM (main.go)..."
    go build -o "../../$OUTPUT_DIR/$PROGRAM" main.go

    # Verify the build
    if [ $? -eq 0 ]; then
        echo "✅ Build successful: $OUTPUT_DIR/$PROGRAM"
    else
        echo "❌ Build failed for $PROGRAM."
        exit 1
    fi

    # Return to the root directory
    cd - > /dev/null
done

echo "🎉 All builds completed successfully. Binaries are in $OUTPUT_DIR"