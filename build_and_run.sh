#!/bin/bash

# Quick build and run script for AIDS

echo "🔨 Building AIDS..."
go build -o aids .

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo "🚀 Running AIDS in real-time mode..."
    ./aids
else
    echo "❌ Build failed!"
    exit 1
fi
