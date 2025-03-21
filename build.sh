#!/bin/bash

export GOOS=darwin
export GOARCH=arm64

OUTPUT_DIR="./app"
mkdir -p "$OUTPUT_DIR"

PROGRAMS=("extract_metrics" "metrics_loadgen")
PROGRAM_PATHS=("src/extract_metrics" "src/metrics_loadgen")
MAIN_FILES=("extract_metrics_main.go" "metrics_loadgen_main.go")

for i in "${!PROGRAMS[@]}"; do
    PROGRAM="${PROGRAMS[$i]}"
    PROGRAM_DIR="${PROGRAM_PATHS[$i]}"
    MAIN_FILE="${MAIN_FILES[$i]}"
    
    echo "🔧 Changing directory to ${PROGRAM_DIR}"
    cd "$PROGRAM_DIR" || { echo "❌ Failed to change directory to $PROGRAM_DIR"; exit 1; }

    echo "📦 Building $PROGRAM (starting from $MAIN_FILE)..."

    go build -o "../../$OUTPUT_DIR/$PROGRAM" "$MAIN_FILE"

    if [ $? -eq 0 ]; then
        echo "✅ Build successful: $OUTPUT_DIR/$PROGRAM"
    else
        echo "❌ Build failed for $PROGRAM."
        exit 1
    fi

    cd - > /dev/null
done

echo "🎉 All builds completed successfully. Binaries are in $OUTPUT_DIR"