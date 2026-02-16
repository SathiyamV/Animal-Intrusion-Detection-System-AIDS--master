#!/bin/bash

# AIDS - Animal Intrusion Detection System - Setup Script
# This script downloads all required model files

set -e  # Exit on error

PROJECT_DIR="/home/parthiban/Downloads/Animal-Intrusion-Detection-System-AIDS--master"
MODEL_DIR="$PROJECT_DIR/model"
ASSET_DIR="$PROJECT_DIR/asset"

echo "=========================================="
echo "AIDS - Setup Script"
echo "=========================================="
echo ""

# Create directories
mkdir -p "$MODEL_DIR"
mkdir -p "$ASSET_DIR"

echo "[1/4] Checking directories..."
echo "  Model directory: $MODEL_DIR"
echo "  Asset directory: $ASSET_DIR"
echo ""

# Download YOLOv4 weights
echo "[2/4] Downloading YOLOv4 weights (237MB)..."
echo "  This may take several minutes depending on your internet connection..."

if [ -f "$MODEL_DIR/yolov4.weights" ]; then
    echo "  ✓ yolov4.weights already exists, skipping..."
else
    cd "$MODEL_DIR"
    wget -q --show-progress https://github.com/AlexeyAB/darknet/releases/download/darknet_yolo_v3_optimal/yolov4.weights || \
    curl -L -o yolov4.weights https://github.com/AlexeyAB/darknet/releases/download/darknet_yolo_v3_optimal/yolov4.weights
    echo "  ✓ Downloaded yolov4.weights"
fi
echo ""

# Download YOLOv4 configuration
echo "[3/4] Downloading YOLOv4 configuration files..."
cd "$MODEL_DIR"

if [ ! -f "$MODEL_DIR/yolov4.cfg" ]; then
    wget -q https://raw.githubusercontent.com/AlexeyAB/darknet/master/cfg/yolov4.cfg || \
    curl -o yolov4.cfg https://raw.githubusercontent.com/AlexeyAB/darknet/master/cfg/yolov4.cfg
    echo "  ✓ Downloaded yolov4.cfg"
else
    echo "  ✓ yolov4.cfg already exists"
fi

if [ ! -f "$MODEL_DIR/coco.names" ]; then
    wget -q https://raw.githubusercontent.com/AlexeyAB/darknet/master/data/coco.names || \
    curl -o coco.names https://raw.githubusercontent.com/AlexeyAB/darknet/master/data/coco.names
    echo "  ✓ Downloaded coco.names"
else
    echo "  ✓ coco.names already exists"
fi
echo ""

# Download sample video if needed
echo "[4/4] Setting up sample video (Optional)..."
if [ ! -f "$ASSET_DIR/traffic.mp4" ]; then
    echo "  No video file found. You can:"
    echo "  1. Place your own video at: $ASSET_DIR/traffic.mp4"
    echo "  2. Download a sample video (requires curl/wget)"
    echo ""
    read -p "  Would you like to download a sample video? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cd "$ASSET_DIR"
        echo "  Downloading sample video..."
        wget -q --show-progress https://www.sample-videos.com/video321/mp4/720/big_buck_bunny_720p_1mb.mp4 -O traffic.mp4 || \
        curl -L -o traffic.mp4 https://www.sample-videos.com/video321/mp4/720/big_buck_bunny_720p_1mb.mp4
        echo "  ✓ Sample video downloaded"
    fi
else
    echo "  ✓ Video file already exists"
fi
echo ""

echo "=========================================="
echo "Setup Complete!"
echo "=========================================="
echo ""
echo "Files created:"
ls -lh "$MODEL_DIR"
echo ""
echo "Next steps:"
echo "1. Change to project directory:"
echo "   cd $PROJECT_DIR"
echo ""
echo "2. Set environment variable:"
echo "   export PKG_CONFIG_PATH=/usr/lib64/pkgconfig:\$PKG_CONFIG_PATH"
echo ""
echo "3. Run the application:"
echo "   ./aids"
echo ""
echo "For more information, see SETUP_GUIDE.md"
