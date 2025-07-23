package config

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	slogbetterstack "github.com/samber/slog-betterstack"
)

type Eurocore struct {
	Url      string `yaml:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Telegram struct {
	Template string `yaml:"template"`
	Id       string `yaml:"id"`
	Secret   string `yaml:"secret"`
	Author   string `yaml:"author"`
}

type Webhook struct {
	Id    string `yaml:"id"`
	Token string `yaml:"token"`
}

type Log struct {
	Level    string `yaml:"level"`
	Token    string `yaml:"token"`
	Endpoint string `yaml:"endpoint"`
}

type Config struct {
	User         string   `yaml:"user"`
	Region       string   `yaml:"region"`
	Limit        uint8    `yaml:"limit"`
	Eurocore     Eurocore `yaml:"eurocore"`
	Webhook      Webhook  `yaml:"webhook"`
	MoveMessage  string   `yaml:"move-message"`
	JoinMessage  string   `yaml:"join-message"`
	MoveTelegram Telegram `yaml:"move-telegram"`
	JoinTelegram Telegram `yaml:"join-telegram"`
	Log          Log      `yaml:"log"`
}

func (c *Config) initLogger() {
	var logger *slog.Logger
	var logLevel slog.Level

	switch c.Log.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	if c.Log.Token != "" && c.Log.Endpoint != "" {
		logger = slog.New(slogbetterstack.Option{
			Token:    c.Log.Token,
			Endpoint: c.Log.Endpoint,
			Level:    logLevel,
		}.NewBetterstackHandler())
	} else {
		logger = slog.Default()
	}

	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(logLevel)
}

func Read() (*Config, error) {
	var path string

	if len(os.Args) > 1 {
		path = os.Args[1]
	} else {
		path = "config.yml"
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return nil, err
	}

	if config.User == "" {
		return nil, errors.New("user not set")
	}

	if config.Region == "" {
		return nil, errors.New("region not set")
	}

	if config.Limit > 50 {
		config.Limit = 50
	}

	config.Region = strings.ReplaceAll(strings.ToLower(config.Region), " ", "_")
	config.Log.Level = strings.ToLower(config.Log.Level)

	if config.MoveMessage == "" {
		config.MoveMessage = "$nation (moved to region)"
	}

	if config.JoinMessage == "" {
		config.JoinMessage = "$nation (joined WA)"
	}

	return config, nil
}
