package storm

import (
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
)

type Config struct {
	Registry struct{
		Url string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"registry"`
	ServerAuthToken string `json:"server_auth_token"`
}

func parseConfig(path string) (*Config, error) {
	cfg := &Config{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Registry: struct {
			Url      string `json:"url"`
			Username string `json:"username"`
			Password string `json:"password"`
		}{Url: "localhost:5000", Username: "username", Password: "password"},
		ServerAuthToken: base64.StdEncoding.EncodeToString([]byte(uuid.New().String())),
	}
}

func createDefaultConfig(path string) error {
	cfg := defaultConfig()
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, os.ModePerm)
}

func InitDefaultConfig(path string) error {
	return createDefaultConfig(path)
}