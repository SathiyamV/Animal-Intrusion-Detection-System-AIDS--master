package engine

import (
	"AIDS/config"
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type Detection struct {
	Box       image.Rectangle
	Label     string
	Timestamp time.Time
	Confidence float32
}

type Detector struct {
	net              gocv.Net
	outputNames      []string
	footage          *gocv.VideoCapture
	window           *gocv.Window
	classes          []string
	config           *config.Config
	lastDetections   map[string]time.Time // Track last detection timestamp for each class
	detectionCooldown time.Duration        // Minimum time between logging same detection
	trackingBoxes    []Detection           // Track detected boxes to keep visible for 1 second
	animalFilter     map[string]bool       // Only show animals, filter out person/train/bench/etc
}

func InitializeDetector(config *config.Config) *Detector {
	return &Detector{
		config:            config,
		lastDetections:    make(map[string]time.Time),
		detectionCooldown: 2 * time.Second,
		trackingBoxes:     []Detection{},
		animalFilter: map[string]bool{
			"cat":       true,
			"dog":       true,
			"horse":     true,
			"sheep":     true,
			"cow":       true,
			"elephant":  true,
			"bear":      true,
			"zebra":     true,
			"giraffe":   true,
			"teddy bear": true,
		},
	}
}

func (d *Detector) Load() error {

	var err error

	d.net = gocv.ReadNet(d.config.Model, d.config.Cfg)

	d.net.SetPreferableBackend(gocv.NetBackendType(gocv.NetBackendDefault))
	d.net.SetPreferableTarget(gocv.NetTargetType(gocv.NetTargetCPU))

	d.outputNames = getOutputsNames(&d.net)

	d.footage, err = gocv.VideoCaptureFile(d.config.Feed)

	if err != nil {
		log.Error(err)
		return err
	}

	d.window = gocv.NewWindow("Animal Intrusion Detection System")

	d.classes = readCOCO(d.config.Classnames)

	return nil
}

func (d *Detector) Process() {

	mat := gocv.NewMat()
	frameCount := 0
	skipFrames := 2  // Process every 3rd frame for real-time performance (skip 2 frames)

	for {
		isTrue := d.footage.Read(&mat)

		if mat.Empty() {
			continue
		}

		if isTrue {
			frameCount++
			
			// Skip frames for real-time performance (process every 3rd frame)
			if frameCount%int(skipFrames+1) == 0 {
				frame, detectedClasses, boxes := detect(&d.net, mat.Clone(), d.config.ScoreThreshold,
					d.config.NmsThreshold, d.outputNames, d.classes, d.animalFilter)
				
				// Update tracking boxes with new detections
				d.trackingBoxes = append(d.trackingBoxes, boxes...)
				
				// Log detections only once per cooldown period to prevent spam
				if len(detectedClasses) > 0 {
					d.logUniqueDections(detectedClasses, frameCount)
				}
				
				d.window.IMShow(frame)
			} else {
				// Display frame with tracked boxes from previous detection
				displayFrame := mat.Clone()
				d.drawTrackedBoxes(&displayFrame)
				d.window.IMShow(displayFrame)
			}
			
			key := d.window.WaitKey(1)
			if key == 113 {
				break
			}
		} else {
			return
		}

	}
}

func (d *Detector) Close() {
	d.net.Close()
	d.footage.Close()
	d.window.Close()
	log.Info("Process Completed")
}

// logUniqueDections prevents spam by only logging each detection once per cooldown period
func (d *Detector) logUniqueDections(detectedClasses []string, frameCount int) {
	now := time.Now()
	for _, class := range detectedClasses {
		lastTime, exists := d.lastDetections[class]
		if !exists || now.Sub(lastTime) > d.detectionCooldown {
			log.Infof("🎯 DETECTED: %s (Frame %d)", class, frameCount)
			d.lastDetections[class] = now
			// Trigger alert sound on new detection
			d.playAlert()
			// Send mobile push notification
			d.sendMobileAlert(class, frameCount)
		}
	}
}
func (d *Detector) playAlert() {
	// Trigger local system beep
	go func() {
		exec.Command("beep", "-f", "1000", "-l", "200").Run()
	}()
}

// sendMobileAlert sends push notification to mobile device via ntfy.sh
func (d *Detector) sendMobileAlert(animal string, frameCount int) {
	go func() {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		// Create message without escape sequences
		message := fmt.Sprintf("[ELEPHANT ALERT] %s detected at Frame %d | Time: %s", animal, frameCount, timestamp)
		
		// Using ntfy.sh for free push notifications
		// User must subscribe: https://ntfy.sh/aids_alerts
		url := "https://ntfy.sh/aids_alerts"
		
		req, _ := http.NewRequest("POST", url, bytes.NewBufferString(message))
		req.Header.Set("Title", fmt.Sprintf("[ALERT] %s Detected!", animal))
		req.Header.Set("Priority", "high")
		
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Warnf("❌ Mobile alert failed: %v", err)
			return
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == 200 {
			log.Infof("✅ Mobile notification sent: %s", animal)
		} else {
			log.Warnf("⚠️ Notification API returned: %d", resp.StatusCode)
		}
	}()
}

// drawTrackedBoxes renders boxes from previous detections that are still within the visibility window
func (d *Detector) drawTrackedBoxes(img *gocv.Mat) {
	now := time.Now()
	const boxVisibilityDuration = 1 * time.Second // Keep boxes visible for 1 second
	
	for _, detection := range d.trackingBoxes {
		// Only draw if box is still within visibility window
		if now.Sub(detection.Timestamp) < boxVisibilityDuration {
			// Draw colored rectangle
			gocv.Rectangle(img, detection.Box, color.RGBA{0, 255, 0, 0}, 2)
			
			// Draw label background
			labelBox := image.Rect(detection.Box.Min.X, detection.Box.Min.Y-25, 
				detection.Box.Min.X+120, detection.Box.Min.Y-5)
			gocv.Rectangle(img, labelBox, color.RGBA{0, 255, 0, 0}, -1)
			
			// Draw text
			gocv.PutText(img, detection.Label, image.Point{detection.Box.Min.X + 3, detection.Box.Min.Y - 8}, 
				gocv.FontHersheySimplex, 0.6, color.RGBA{0, 0, 0, 0}, 1)
		}
	}
	
	// Clean up old boxes
	var activeBoxes []Detection
	for _, detection := range d.trackingBoxes {
		if now.Sub(detection.Timestamp) < boxVisibilityDuration {
			activeBoxes = append(activeBoxes, detection)
		}
	}
	d.trackingBoxes = activeBoxes
}

func detect(net *gocv.Net, src gocv.Mat, scoreThreshold float32, nmsThreshold float32, OutputNames []string, classes []string, animalFilter map[string]bool) (gocv.Mat, []string, []Detection) {
	img := src.Clone()
	img.ConvertTo(&img, gocv.MatTypeCV32F)
	blob := gocv.BlobFromImage(img, 1/255.0, image.Pt(416, 416), gocv.NewScalar(0, 0, 0, 0), true, false)
	net.SetInput(blob, "")
	probs := net.ForwardLayers(OutputNames)
	boxes, confidences, classIds := postProcess(img, &probs, classes, animalFilter)
	indices := make([]int, 100)
	if len(boxes) == 0 { // No Classes
		return src, []string{}, []Detection{}
	}
	gocv.NMSBoxes(boxes, confidences, scoreThreshold, nmsThreshold, indices)

	return drawRect(src, boxes, classes, classIds, indices)
}

func postProcess(frame gocv.Mat, outs *[]gocv.Mat, classes []string, animalFilter map[string]bool) ([]image.Rectangle, []float32, []int) {
	var classIds []int
	var confidences []float32
	var boxes []image.Rectangle
	for _, out := range *outs {

		data, _ := out.DataPtrFloat32()
		for i := 0; i < out.Rows(); i, data = i+1, data[out.Cols():] {

			scoresCol := out.RowRange(i, i+1)

			scores := scoresCol.ColRange(5, out.Cols())
			_, confidence, _, classIDPoint := gocv.MinMaxLoc(scores)
			// Lowered threshold to 0.35 for better elephant detection in night vision
			if confidence > 0.35 && classIDPoint.X >= 0 && classIDPoint.X < len(classes) {
				classLabel := classes[classIDPoint.X]
				if animalFilter[classLabel] { // Only include if it's an animal
					centerX := int(data[0] * float32(frame.Cols()))
					centerY := int(data[1] * float32(frame.Rows()))
					width := int(data[2] * float32(frame.Cols()))
					height := int(data[3] * float32(frame.Rows()))

					left := centerX - width/2
					top := centerY - height/2
					right := left + width
					bottom := top + height
					classIds = append(classIds, classIDPoint.X)
					confidences = append(confidences, float32(confidence))
					boxes = append(boxes, image.Rect(left, top, right, bottom))
				}
			}
		}
	}
	return boxes, confidences, classIds
}

func drawRect(img gocv.Mat, boxes []image.Rectangle, classes []string, classIds []int, indices []int) (gocv.Mat, []string, []Detection) {
	var detectClass []string
	var detections []Detection
	now := time.Now()
	
	for _, idx := range indices {
		if idx == 0 || idx >= len(boxes) {
			continue
		}
		
		// Get box coordinates
		x := boxes[idx].Min.X
		y := boxes[idx].Min.Y
		w := boxes[idx].Max.X
		h := boxes[idx].Max.Y
		
		// Get label
		if classIds[idx] >= 0 && classIds[idx] < len(classes) {
			label := classes[classIds[idx]]
			
			// For this elephant detection system: convert horses to elephants
			// Since elephants are often misclassified as horses by YOLO, especially at distance
			if label == "horse" {
				label = "elephant"
			}
			
			// Draw colored rectangle
			gocv.Rectangle(&img, image.Rect(x, y, w, h), color.RGBA{0, 255, 0, 0}, 2)
			
			// Draw label background (filled rectangle)
			gocv.Rectangle(&img, image.Rect(x, y-25, x+120, y-5), color.RGBA{0, 255, 0, 0}, -1)
			
			// Draw text (label name in black)
			gocv.PutText(&img, label, image.Point{x + 3, y - 8}, gocv.FontHersheySimplex, 0.6, color.RGBA{0, 0, 0, 0}, 1)
			
			detectClass = append(detectClass, label)
			
			// Track detection for persistence (visible for 1 second)
			detections = append(detections, Detection{
				Box:       image.Rect(x, y, w, h),
				Label:     label,
				Timestamp: now,
			})
		}
	}
	return img, detectClass, detections
}

func getOutputsNames(net *gocv.Net) []string {
	var outputLayers []string
	for _, i := range net.GetUnconnectedOutLayers() {
		layer := net.GetLayer(i)
		layerName := layer.GetName()
		if layerName != "_input" {
			outputLayers = append(outputLayers, layerName)
		}
	}
	return outputLayers
}

func readCOCO(path string) []string {
	var classes []string
	read, _ := os.Open(path)
	defer read.Close()
	for {
		var t string
		_, err := fmt.Fscan(read, &t)
		if err != nil {
			break
		}
		classes = append(classes, t)
	}
	return classes
}
