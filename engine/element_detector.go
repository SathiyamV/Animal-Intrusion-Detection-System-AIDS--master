package engine

import (
	"AIDS/config"
	"bufio"
	"fmt"
	"image"
	"image/color"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

type Detection struct {
	Box        image.Rectangle
	Label      string
	Timestamp  time.Time
	Confidence float32
}

type Detector struct {
	net               gocv.Net
	outputNames       []string
	footage           *gocv.VideoCapture
	window            *gocv.Window
	classes           []string
	config            *config.Config
	lastDetections    map[string]time.Time // Track last detection timestamp for each class
	detectionCooldown time.Duration        // Minimum time between logging same detection
	lastNotifications map[string]time.Time // Track last notification timestamp for each class
	notifyCooldown    time.Duration        // Minimum time between notifications for same class
	trackingBoxes     []Detection          // Track detected boxes to keep visible for 1 second
	animalFilter      map[string]bool      // Only show animals, filter out person/train/bench/etc
	telegramWarned    bool                 // Avoid repeating setup warnings when Telegram config is missing
}

func InitializeDetector(config *config.Config) *Detector {
	detectionCooldown := 2 * time.Second
	if config.DetectionLogCooldownMilli > 0 {
		detectionCooldown = time.Duration(config.DetectionLogCooldownMilli) * time.Millisecond
	}

	notifyCooldown := 30 * time.Second
	if config.NotificationCooldownSec > 0 {
		notifyCooldown = time.Duration(config.NotificationCooldownSec) * time.Second
	}

	return &Detector{
		config:            config,
		lastDetections:    make(map[string]time.Time),
		detectionCooldown: detectionCooldown,
		lastNotifications: make(map[string]time.Time),
		notifyCooldown:    notifyCooldown,
		trackingBoxes:     []Detection{},
		animalFilter: map[string]bool{
			"cat":        true,
			"dog":        true,
			"horse":      true,
			"sheep":      true,
			"cow":        true,
			"elephant":   true,
			"bear":       true,
			"zebra":      true,
			"giraffe":    true,
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
	log.Infof("Loaded %d classes", len(d.classes))
	log.Infof("Class index dog=%d cow=%d", indexOf(d.classes, "dog"), indexOf(d.classes, "cow"))

	return nil
}

func (d *Detector) Process() {

	mat := gocv.NewMat()
	defer mat.Close()

	frameCount := 0
	skipFrames := 2 // Default: process every 3rd frame
	if d.config.FrameSkip > 0 {
		skipFrames = d.config.FrameSkip
	}

	for {
		shouldStop := false
		feedEnded := false

		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("Recovered from frame processing panic: %v", r)
				}
			}()

			isTrue := d.footage.Read(&mat)

			if mat.Empty() {
				return
			}

			if isTrue {
				frameCount++

				// Skip frames for real-time performance (process every 3rd frame)
				if frameCount%int(skipFrames+1) == 0 {
					frame, detectedClasses, boxes := detect(&d.net, mat.Clone(), d.config.ScoreThreshold,
						d.config.NmsThreshold, d.outputNames, d.classes, d.animalFilter)
					defer frame.Close()

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
					defer displayFrame.Close()
					d.drawTrackedBoxes(&displayFrame)
					d.window.IMShow(displayFrame)
				}

				key := d.window.WaitKey(1)
				if key == 113 {
					shouldStop = true
				}
			} else {
				feedEnded = true
			}
		}()

		if shouldStop || feedEnded {
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
			// Send Telegram notification with stronger per-animal cooldown
			d.sendTelegramAlert(class, frameCount, now)
		}
	}
}
func (d *Detector) playAlert() {
	// Trigger local system beep
	go func() {
		exec.Command("beep", "-f", "1000", "-l", "200").Run()
	}()
}

// sendTelegramAlert sends notification to Telegram chat using Bot API.
func (d *Detector) sendTelegramAlert(animal string, frameCount int, now time.Time) {
	if !d.config.EnableTelegram {
		return
	}

	lastNotification, exists := d.lastNotifications[animal]
	if exists && now.Sub(lastNotification) < d.notifyCooldown {
		return
	}
	d.lastNotifications[animal] = now

	if d.config.TelegramBotToken == "" || d.config.TelegramChatID == "" {
		if !d.telegramWarned {
			log.Warn("Telegram enabled but not configured: set telegramBotToken and telegramChatId in config/aids.cfg")
			d.telegramWarned = true
		}
		return
	}

	go func() {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message := fmt.Sprintf("AIDS Alert: %s detected at frame %d (%s)", animal, frameCount, timestamp)
		botURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", d.config.TelegramBotToken)
		form := url.Values{
			"chat_id":    {d.config.TelegramChatID},
			"text":       {message},
			"parse_mode": {"HTML"},
		}

		req, err := http.NewRequest("POST", botURL, strings.NewReader(form.Encode()))
		if err != nil {
			log.Warnf("Telegram request creation failed: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Warnf("Telegram alert failed: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			log.Infof("Telegram notification sent: %s", animal)
		} else {
			log.Warnf("Telegram API returned status: %d", resp.StatusCode)
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
	defer img.Close()

	img.ConvertTo(&img, gocv.MatTypeCV32F)
	blob := gocv.BlobFromImage(img, 1/255.0, image.Pt(416, 416), gocv.NewScalar(0, 0, 0, 0), true, false)
	defer blob.Close()

	net.SetInput(blob, "")
	probs := net.ForwardLayers(OutputNames)
	defer func() {
		for i := range probs {
			probs[i].Close()
		}
	}()

	boxes, confidences, classIds := postProcess(img, &probs, classes, animalFilter, scoreThreshold)
	if len(boxes) == 0 { // No Classes
		return src, []string{}, []Detection{}
	}
	indices := make([]int, len(boxes))
	for i := range indices {
		indices[i] = -1
	}
	gocv.NMSBoxes(boxes, confidences, scoreThreshold, nmsThreshold, indices)

	return drawRect(src, boxes, classes, classIds, indices)
}

func postProcess(frame gocv.Mat, outs *[]gocv.Mat, classes []string, animalFilter map[string]bool, scoreThreshold float32) ([]image.Rectangle, []float32, []int) {
	var classIds []int
	var confidences []float32
	var boxes []image.Rectangle
	for _, out := range *outs {

		data, _ := out.DataPtrFloat32()
		for i := 0; i < out.Rows(); i, data = i+1, data[out.Cols():] {

			scoresCol := out.RowRange(i, i+1)

			scores := scoresCol.ColRange(5, out.Cols())
			_, classConfidence, _, classIDPoint := gocv.MinMaxLoc(scores)
			objectness := data[4]
			confidence := float32(classConfidence) * objectness
			if confidence > scoreThreshold && classIDPoint.X >= 0 && classIDPoint.X < len(classes) {
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
		if idx < 0 || idx >= len(boxes) || idx >= len(classIds) {
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
			if label == "dog" || label == "cow" {
				log.Infof("Detection classId=%d label=%s", classIds[idx], label)
			}

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

	scanner := bufio.NewScanner(read)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		classes = append(classes, line)
	}
	return classes
}

func indexOf(classes []string, target string) int {
	for i, c := range classes {
		if c == target {
			return i
		}
	}
	return -1
}
