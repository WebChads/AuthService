package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Port          string         `json:"port" env:"PORT"`
	SecretKey     string         `json:"secret_key" env:"SECRET_KEY"`
	IsDevelopment bool           `json:"is_development" env:"IS_DEVELOPMENT"`
	DbSettings    DatabaseConfig `json:"database"`
	KafkaConfig   KafkaConfig    `json:"kafka"`
}

type DatabaseConfig struct {
	Host     string `json:"host" env:"DATABASE__HOST"`
	DbName   string `json:"db_name" env:"DATABASE__DB_NAME"`
	User     string `json:"user" env:"DATABASE__USER"`
	Password string `json:"password" env:"DATABASE__PASSWORD"`
}

type KafkaConfig struct {
	Url string `json:"url" env:"KAFKA__URL"`
}

var cfg AppConfig
var cachedProjectRootPath string

func InitializeConfig() (*AppConfig, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	projectRootPath, err := findProjectRoot(workingDirectory)
	if err != nil {
		return nil, err
	}

	path := projectRootPath + "/configs/appsettings.json"

	// Reading config from file
	err = cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Then overriding with environment variables
	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read env vars: %w", err)
	}

	// Validate required fields
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validateConfig(cfg *AppConfig) error {
	var missing []string

	if cfg.Port == "" {
		missing = append(missing, "port")
	}
	if cfg.SecretKey == "" {
		missing = append(missing, "secret_key")
	}
	if cfg.DbSettings.Host == "" {
		missing = append(missing, "database.host")
	}
	if cfg.DbSettings.DbName == "" {
		missing = append(missing, "database.db_name")
	}
	if cfg.DbSettings.User == "" {
		missing = append(missing, "database.user")
	}
	if cfg.DbSettings.Password == "" {
		missing = append(missing, "database.password")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required config fields: %s", strings.Join(missing, ", "))
	}

	return nil
}

func findProjectRoot(startDir string) (string, error) {
	if cachedProjectRootPath != "" {
		return cachedProjectRootPath, nil
	}

	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			cachedProjectRootPath = dir
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("project root not found")
		}
		dir = parent
	}
}
