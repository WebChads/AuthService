package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"log/slog"

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
	Host     string `json:"host" env:"DATABASE_HOST"`
	DbName   string `json:"db_name" env:"DATABASE_DB_NAME"`
	User     string `json:"user" env:"DATABASE_USER"`
	Password string `json:"password" env:"DATABASE_PASSWORD"`
}

type KafkaConfig struct {
	Url string `json:"url" env:"KAFKA_URL"`
}

var cfg AppConfig
var cachedProjectRootPath string

// Initialize config - from first from config file, then from env vars (overriding). So you can choose any variant you want (even mixing)
func InitializeConfig() (*AppConfig, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Config from appsetting.json (variant through config file)
	projectRootPath, err := findProjectRoot(workingDirectory)
	if err != nil {
		fmt.Println("didn't found project root (it depends on go.mod file) - if you're using env variables - skip this message")
	} else {
		path := projectRootPath + "/configs/appsettings.json"

		// Reading config from file
		err = cleanenv.ReadConfig(path, &cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Config from environment variables (variant through env)
	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		fmt.Println("can't read env variables - if you're using config file - skip this message")
	}

	// Validate required fields
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	fmt.Sprintf("%v\n", cfg)
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
