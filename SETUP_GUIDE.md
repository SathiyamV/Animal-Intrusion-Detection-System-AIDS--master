# Animal Intrusion Detection System (AIDS) - Setup Guide

## Build Status ✅

The application has been successfully built and compiled as `aids` binary.

## Prerequisites Installed

- ✅ Go 1.25.7
- ✅ OpenCV 4.11.0 with development headers
- ✅ C++ Compiler (gcc-c++)
- ✅ Go dependencies (logrus, gocv)

## Required Model Files

The application requires YOLO v4 model files to function. Download the following files and place them in the `model/` directory:

### 1. YOLOv4 Weights (237 MB)

```
File: yolov4.weights
Source: https://github.com/AlexeyAB/darknet/releases/download/darknet_yolo_v3_optimal/yolov4.weights
Location: ./model/yolov4.weights
```

### 2. YOLOv4 Configuration (167 KB)

```
File: yolov4.cfg
Source: https://raw.githubusercontent.com/AlexeyAB/darknet/master/cfg/yolov4.cfg
Location: ./model/yolov4.cfg
```

### 3. COCO Class Names (625 B)

```
File: coco.names
Source: https://raw.githubusercontent.com/AlexeyAB/darknet/master/data/coco.names
Location: ./model/coco.names
```

### 4. Sample Video File (ANY FORMAT)

```
Location: ./asset/traffic.mp4
Supported formats: .mp4, .avi, .mov, .flv, etc.
```

## Download Instructions

### Option 1: Manual Download (Recommended for slow connections)

```bash
cd /home/parthiban/Downloads/Animal-Intrusion-Detection-System-AIDS--master

# Create directories if they don't exist
mkdir -p model asset

# Download YOLO v4 weights (this is a large file ~237MB)
cd model
wget https://github.com/AlexeyAB/darknet/releases/download/darknet_yolo_v3_optimal/yolov4.weights

# Download configuration files
wget https://raw.githubusercontent.com/AlexeyAB/darknet/master/cfg/yolov4.cfg
wget https://raw.githubusercontent.com/AlexeyAB/darknet/master/data/coco.names

cd ../asset
# Download a sample video (or use your own)
# You can use any video file. For testing, you can use:
wget https://www.sample-videos.com/video321/mp4/720/big_buck_bunny_720p_1mb.mp4 -O traffic.mp4
```

### Option 2: Using curl

```bash
cd /home/parthiban/Downloads/Animal-Intrusion-Detection-System-AIDS--master/model

# Download weights
curl -L -o yolov4.weights https://github.com/AlexeyAB/darknet/releases/download/darknet_yolo_v3_optimal/yolov4.weights

# Download config and names
curl -o yolov4.cfg https://raw.githubusercontent.com/AlexeyAB/darknet/master/cfg/yolov4.cfg
curl -o coco.names https://raw.githubusercontent.com/AlexeyAB/darknet/master/data/coco.names
```

## Running the Application

### Method 1: Direct Execution

```bash
cd /home/parthiban/Downloads/Animal-Intrusion-Detection-System-AIDS--master

# Set PKG_CONFIG_PATH for OpenCV
export PKG_CONFIG_PATH=/usr/lib64/pkgconfig:$PKG_CONFIG_PATH

# Run the application
./aids
```

### Method 2: Using Go Run

```bash
cd /home/parthiban/Downloads/Animal-Intrusion-Detection-System-AIDS--master

# Set PKG_CONFIG_PATH and run
export PKG_CONFIG_PATH=/usr/lib64/pkgconfig:$PKG_CONFIG_PATH
go run .
```

### Method 3: Create an Alias (Optional)

```bash
# Add this to your ~/.bashrc or ~/.zshrc
alias aids='cd /home/parthiban/Downloads/Animal-Intrusion-Detection-System-AIDS--master && export PKG_CONFIG_PATH=/usr/lib64/pkgconfig:$PKG_CONFIG_PATH && ./aids'

# Then run it anytime with:
aids
```

## Configuration

Modify `config/aids.cfg` to customize the detection system:

```json
{
  "model": "./model/yolov4.weights",
  "cfg": "./model/yolov4.cfg",
  "feed": "./asset/traffic.mp4",
  "classnames": "./model/coco.names",
  "gpuEnabled": false,
  "scoreThreshold": 0.45,
  "nmsThreshold": 0.5
}
```

### Configuration Parameters:

- **model**: Path to YOLOv4 weights file
- **cfg**: Path to YOLOv4 configuration file
- **feed**: Path to video file or video capture device (0 for webcam)
- **classnames**: Path to class names file
- **gpuEnabled**: Set to true if GPU CUDA is available
- **scoreThreshold**: Confidence threshold for detections (0.0-1.0)
- **nmsThreshold**: Non-Maximum Suppression threshold (0.0-1.0)

## Rebuilding the Application

If you make changes to the code, rebuild with:

```bash
cd /home/parthiban/Downloads/Animal-Intrusion-Detection-System-AIDS--master
export PKG_CONFIG_PATH=/usr/lib64/pkgconfig:$PKG_CONFIG_PATH
go build -o aids .
```

## Troubleshooting

### "Config file not found" Error

Make sure you're running the application from the project root directory.

### "Model file not found" Error

- Verify model files exist in `./model/` directory
- Check file permissions: `ls -la model/`

### "Video file cannot be opened" Error

- Verify video file path in config
- Try using absolute path: `/home/user/path/to/video.mp4`
- Test with a different video file

### OpenCV pkg-config errors

```bash
export PKG_CONFIG_PATH=/usr/lib64/pkgconfig:$PKG_CONFIG_PATH
```

### GPU Not Working

Either disable GPU in config (`"gpuEnabled": false`) or install NVIDIA CUDA toolkit.

## System Requirements

- **RAM**: Minimum 2GB (4GB+ recommended for smooth operation)
- **Disk**: 250GB+ free space (for model files)
- **Processor**: Multi-core processor recommended
- **Video formats**: MP4, AVI, MOV, FLV, etc.

## Next Steps

1. Download the model files (see Download Instructions above)
2. Prepare or download a video file
3. Update `config/aids.cfg` with correct file paths
4. Run the application using one of the methods above

## Support

For issues or questions about the Animals Intrusion Detection System, refer to the main README.md
