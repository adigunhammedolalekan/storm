package storm

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Registry struct{
		Url string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
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
		ServerAuthToken: "",
	}
}
