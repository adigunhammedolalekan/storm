package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
)

type StormClient struct {
	cmd *CmdClient
	httpClient *http.Client
	config *Config
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

func (s *StormClient) BuildBinary() error {
	return s.cmd.ExecBuildCommand()
}

// env GOOS=linux go build -ldflags="-s -w" -o stormTest main.go
type CmdClient struct {
	appName string
}

func NewCmdClient(appName string) *CmdClient {
	return &CmdClient{appName:appName}
}

func (c *CmdClient) ExecBuildCommand() error {
	export := exec.Command("env", "GOOS=linux")
	if err := export.Run(); err != nil {
		return err
	}
	if err := exec.Command("echo", "$GOOS").Run(); err != nil {
		return err
	}
	// cmd := exec.Command("go", "build", "-o", c.appName)
	cmd := exec.Command(`env GOOS=linux go build -ldflags="-s -w" -o ` + c.appName)
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

