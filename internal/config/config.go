package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)


type Config struct {
	Username 		string `json:"username"`
	CheckInterval	string `json:"check_interval"`
}


func Load() (*Config, error) {
	config := &Config{
		CheckInterval: "10s",
	}

	configPath, err := configPath()
	if err != nil {
		return nil, err
	}

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist){
			return config, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(configBytes, config); err != nil {
		return nil, err
	}

	return config, nil
}

func Save(config *Config) error {
	configPath, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	configJson, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, configJson, 0600); err != nil {
		return err
	}
	return nil
}

func (c *Config) Interval() time.Duration {
    d, err := time.ParseDuration(c.CheckInterval)
    if err != nil || d <= 0 {
        return 10 * time.Second
    }
    return d
}




func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	pathString := filepath.Join(dir, "kawaii-wifi", "config.json")
	return pathString, nil
}