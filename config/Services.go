package config

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

type Config struct {
	Model                     string  `json:"model"`
	Cfg                       string  `json:"cfg"`
	Feed                      string  `json:"feed"`
	Classnames                string  `json:"classnames"`
	GpuEnabled                bool    `json:"gpuEnabled"`
	ScoreThreshold            float32 `json:"scoreThreshold"`
	NmsThreshold              float32 `json:"nmsThreshold"`
	EnableTelegram            bool    `json:"enableTelegram"`
	TelegramBotToken          string  `json:"telegramBotToken"`
	TelegramChatID            string  `json:"telegramChatId"`
	NotificationCooldownSec   int     `json:"notificationCooldownSec"`
	DetectionLogCooldownMilli int     `json:"detectionLogCooldownMs"`
	FrameSkip                 int     `json:"frameSkip"`
}

func LoadConfig() (*Config, error) {
	configFile, err := os.OpenFile("./config/aids.cfg", os.O_RDONLY, 0666)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	configBytes, _ := io.ReadAll(configFile)

	var config Config = Config{}

	err = json.Unmarshal(configBytes, &config)
	if err != nil {

		return nil, err
	}

	log.Printf(
		"\nLoaded Config: model=%s feed=%s classnames=%s gpuEnabled=%t scoreThreshold=%.2f nmsThreshold=%.2f enableTelegram=%t telegramConfigured=%t notificationCooldownSec=%d detectionLogCooldownMs=%d frameSkip=%d",
		config.Model,
		config.Feed,
		config.Classnames,
		config.GpuEnabled,
		config.ScoreThreshold,
		config.NmsThreshold,
		config.EnableTelegram,
		config.TelegramBotToken != "" && config.TelegramChatID != "",
		config.NotificationCooldownSec,
		config.DetectionLogCooldownMilli,
		config.FrameSkip,
	)

	return &config, nil
}
