package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
)
// StormClient
type StormClient struct {
	cmd        *CmdClient
	httpClient *http.Client
	config     *Config
}

func NewStormClient(config *Config) (*StormClient, error) {
	if len(config.AppName) == 0 {
		return nil, errors.New("appName is missing")
	}
	_, err := url.Parse(config.ServerUrl)
	if err != nil {
		return nil, err
	}
	if len(config.ServerAuthCode) == 0 {
		return nil, errors.New("server authentication code is missing")
	}
	cmd := NewCmdClient(config.AppName)
	httpClient := &http.Client{Timeout: 60 * time.Second}
	return &StormClient{
		cmd:        cmd,
		httpClient: httpClient,
		config:     config,
	}, nil
}

func (s *StormClient) DeployApp(binPath string, result *DeploymentResult) error {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	if err := writer.WriteField("app_name", s.config.AppName); err != nil {
		return err
	}
	for _, env := range s.config.Environment {
		if err := writer.WriteField(env.Key, env.Value); err != nil {
			log.Println("failed to write env variable: ", err)
		}
	}
	in, err := os.Open(binPath)
	if err != nil {
		return err
	}
	out, err := writer.CreateFormFile("bin", binPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	serverUrl := fmt.Sprintf("%s/deploy", s.config.ServerUrl)
	req, err := http.NewRequest("POST", serverUrl, buf)
	if err != nil {
		return err
	}
	req.Header.Set("X-Server-Code", s.config.ServerAuthCode)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, result); err != nil {
		return err
	}
	if r.StatusCode != http.StatusOK {
		return errors.New(result.Message)
	}
	return nil
}

func (s *StormClient) GetAppLogs() (string, error) {
	u := fmt.Sprintf("%s/logs/%s", s.config.ServerUrl, s.config.AppName)
	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", err
	}
	r.Header.Set("X-Server-Code", s.config.ServerAuthCode)
	response, err := s.httpClient.Do(r)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	var data struct{
		Error bool `json:"error"`
		Message string `json:"message"`
		Data struct{
			Logs string `json:"logs"`
		}
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK || data.Error {
		return "", errors.New(data.Message)
	}
	return data.Data.Logs, nil
}

func (s *StormClient) BuildBinary() error {
	return s.cmd.ExecBuildCommand()
}

// env GOOS=linux go build -ldflags="-s -w" -o stormTest main.go
type CmdClient struct {
	appName string
}

func NewCmdClient(appName string) *CmdClient {
	return &CmdClient{appName: appName}
}

func (c *CmdClient) ExecBuildCommand() error {
	// change build env to linux, we need linux container
	if err := os.Setenv("GOOS", "linux"); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", c.appName)
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	cmd.Dir = wd
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}