package config

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	f, err := os.Open(configFilePath)
	if err != nil {
		// if the file does not exist, create it with default values
		if errors.Is(err, os.ErrNotExist) {
			cfg := Config{
				DbURL: "postgres://example",
			}

			if err := write(cfg); err != nil {
				return Config{}, err
			}

			return cfg, nil
		}

		// any other error, just return
		return Config{}, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg *Config) SetUser(userName string) error {
	cfg.CurrentUserName = userName

	write(*cfg)

	return nil
}

func getConfigFilePath() (string, error) {
	userHomeDirectory, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configFilePath := userHomeDirectory + "/" + configFileName
	return configFilePath, nil

}

func write(cfg Config) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(configFilePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(cfg); err != nil {
		return err
	}

	return nil

}
