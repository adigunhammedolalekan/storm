package storm

import (
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"os"
)

type Config struct {
	Registry struct {
		Url      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"registry"`
	ServerAuthToken string `json:"server_auth_token"`
}

func parseConfig(r io.Reader) (*Config, error) {
	cfg := &Config{}
	if err := json.NewDecoder(r).Decode(cfg); err != nil {
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

func createDefaultConfig(location *os.File) error {
	cfg := defaultConfig()
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}

	if _, err := io.WriteString(location, string(data)); err != nil {
		return err
	}
	return nil
}

func InitDefaultConfig(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	return createDefaultConfig(f)
}
