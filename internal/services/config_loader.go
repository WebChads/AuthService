package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Port          string         `json:"port"`
	SecretKey     string         `json:"secret_key"`
	IsDevelopment bool           `json:"is_development"`
	DbSettings    DatabaseConfig `json:"database"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	DbName   string `json:"db_name"`
	User     string `json:"user"`
	Password string `json:"password"`
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

	path := projectRootPath + "\\configs\\appsettings.json"

	err = cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, err
	}

	if len(cfg.Port) == 0 || len(cfg.SecretKey) == 0 || cfg.DbSettings.Host == "" || cfg.DbSettings.DbName == "" || cfg.DbSettings.User == "" || cfg.DbSettings.Password == "" {
		return nil, errors.New("unsuccess try of reading config")
	}

	return &cfg, nil
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
