package config

import (
	"os"
	"path/filepath"
	"strings"
)

const envDataPath = "MITORI_DATA_PATH"

type Config struct {
	DataPath string
}

func Load() (Config, error) {
	if v := strings.TrimSpace(os.Getenv(envDataPath)); v != "" {
		p, err := expandHome(v)
		if err != nil {
			return Config{}, err
		}
		return Config{DataPath: p}, nil
	}

	p, err := expandHome("~/.config/mitori/data.json")
	if err != nil {
		return Config{}, err
	}
	return Config{DataPath: p}, nil
}

func expandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return filepath.Clean(path), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Clean(filepath.Join(home, strings.TrimPrefix(path, "~/"))), nil
}
