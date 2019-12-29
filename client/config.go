package client

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	ServerUrl      string `json:"server_url"`
	ServerAuthCode string `json:"server_auth_code"`
	AppName        string `json:"app_name"`
}

func Parse(r io.Reader) (*Config, error) {
	cfg := &Config{}
	if err := json.NewDecoder(r).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Generate(cfg *Config) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(wd, "storm_config.json")
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, os.ModePerm)
}