# 🎯 AIDS - Animal Intrusion Detection System - Installation Summary

## ✅ Completed Installation Steps

### 1. System Dependencies Installed

- ✅ **Go 1.25.7** - Programming language
- ✅ **OpenCV 4.11.0** - Computer vision library
- ✅ **C++ Compiler** (gcc-c++ 15.2.1) - Required for GoCV compilation
- ✅ **Build Tools** (make, pkg-config) - Build utilities
- ✅ **Version Control** (git, subversion, mercurial) - Already installed

### 2. Go Module Dependencies Downloaded

- ✅ **github.com/sirupsen/logrus v1.9.0** - Logging library
- ✅ **gocv.io/x/gocv v0.32.1** - Go OpenCV bindings
- ✅ **golang.org/x/sys** - System utilities

### 3. Application Built Successfully

- ✅ **Binary created**: `aids` (5.4 MB)
- ✅ **Build type**: Release
- ✅ **Location**: /home/parthiban/Downloads/Animal-Intrusion-Detection-System-AIDS--master/aids

### 4. Configuration Files Prepared

- ✅ **config/aids.cfg** - Main configuration with YOLO settings
- ✅ **model/coco.names** - 80 COCO dataset class names

## 📋 What You Need To Do Next

### Step 1: Download Model Files (237 MB)

The YOLOv4 neural network model weights are required. You have two options:

**Option A: Automated Download (Recommended)**

```bash
cd /home/parthiban/Downloads/Animal-Intrusion-Detection-System-AIDS--master
bash setup.sh
```

This will interactively download all required files.

**Option B: Manual Downloads**

```bash
cd /home/parthiban/Downloads/Animal-Intrusion-Detection-System-AIDS--master
mkdir -p model asset

# Download YOLOv4 weights (237 MB - takes 1-10 minutes)
wget https://github.com/AlexeyAB/darknet/releases/download/darknet_yolo_v3_optimal/yolov4.weights -P model/

# Download config and class names (small files, fast)
wget https://raw.githubusercontent.com/AlexeyAB/darknet/master/cfg/yolov4.cfg -P model/
wget https://raw.githubusercontent.com/AlexeyAB/darknet/master/data/coco.names -P model/
```

### Step 2: Prepare a Video File

Place a video file at `./asset/traffic.mp4` or update config to point to your video:

- Supported formats: MP4, AVI, MOV, FLV, etc.
- Or use your webcam (change config "feed" to "0")

### Step 3: Run the Application

```bash
cd /home/parthiban/Downloads/Animal-Intrusion-Detection-System-AIDS--master

# Set OpenCV pkg-config path
export PKG_CONFIG_PATH=/usr/lib64/pkgconfig:$PKG_CONFIG_PATH

# Run the application
./aids
```

The application will:

1. Load the YOLOv4 model
2. Open the video feed
3. Detect animals and objects in real-time
4. Display results with bounding boxes
5. Exit when you press 'Q'

## 📂 Project Structure

```
AIDS--master/
├── aids (executable binary) ✅ READY
├── aids.go (main entry point)
├── go.mod (dependencies)
├── config/
│   ├── aids.cfg (configuration)
│   └── Services.go
├── engine/
│   ├── aids_main_engine.go
│   └── element_detector.go
├── model/
│   ├── coco.names ✅ READY
│   ├── yolov4.weights ❌ NEEDS DOWNLOAD (237 MB)
│   └── yolov4.cfg ❌ NEEDS DOWNLOAD
├── asset/
│   └── traffic.mp4 ❌ NEEDS VIDEO FILE
├── SETUP_GUIDE.md (detailed setup guide)
└── setup.sh (automated setup script)
```

## 🚀 Quick Start Commands

```bash
# Navigate to project
cd /home/parthiban/Downloads/Animal-Intrusion-Detection-System-AIDS--master

# Download models automatically
bash setup.sh

# Setup environment and run
export PKG_CONFIG_PATH=/usr/lib64/pkgconfig:$PKG_CONFIG_PATH
./aids
```

## ⚙️ Configuration

Edit `config/aids.cfg` to customize:

```json
{
  "model": "./model/yolov4.weights",
  "cfg": "./model/yolov4.cfg",
  "feed": "./asset/traffic.mp4", // or "0" for webcam
  "classnames": "./model/coco.names",
  "gpuEnabled": false, // set true if GPU available
  "scoreThreshold": 0.45, // detection confidence
  "nmsThreshold": 0.5 // overlap suppression
}
```

## 📊 System Requirements

- **OS**: Linux (Fedora 43 confirmed)
- **RAM**: 2GB+ available
- **Disk**: 250GB+ free space (for models)
- **CPU**: Multi-core processor recommended
- **Internet**: Required for downloading models (one-time)

## 🔧 Troubleshooting

### Missing dquote or syntax error

Terminal state issue - just run commands again

### "Model not found" error

- Verify model files exist: `ls -la model/`
- Check file permissions: `chmod 644 model/*`

### "Cannot open video" error

- Update path in `config/aids.cfg`
- Use absolute path: `/home/user/path/to/video.mp4`

### pkg-config error when rebuilding

```bash
export PKG_CONFIG_PATH=/usr/lib64/pkgconfig:$PKG_CONFIG_PATH
go build -o aids .
```

## 📝 Additional Information

For detailed documentation, see:

- **README.md** - Project overview and research
- **SETUP_GUIDE.md** - Comprehensive setup instructions
- **config/aids.cfg** - Configuration options

## 🎓 What This System Does

The AIDS (Animal Intrusion Detection System) uses:

- **YOLOv4**: State-of-the-art object detection algorithm
- **Go/GoCV**: High-performance video processing
- **OpenCV**: Computer vision operations

It detects and displays bounding boxes around:

- Animals (dogs, cats, cows, elephants, birds, etc.)
- Vehicles (cars, trucks, motorcycles)
- People and other objects

Perfect for:

- Wildlife monitoring
- Agricultural protection
- Urban security
- Surveillance analysis

---

**Status**: ✅ Application ready to run. Just need to download model files!

**Next Action**: Run `bash setup.sh` to download models, then execute `./aids`
