package main

import (
	"errors"
	"fmt"
	"github.com/adigunhammedolalekan/storm/client"
	"github.com/manifoldco/promptui"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

func main() {
	args := os.Args
	if len(args) > 1 {
		arg := args[1]
		switch arg {
		case "init":
			initProjectConfig()
		case "logs":
			getLogs()
		default:
			runClient()
		}
	}
}

type consoleReader struct {
}

func (c *consoleReader) readValue(message string, validate func(string) error) string {
	p := promptui.Prompt{
		Label: message,
		Validate: validate,
	}
	r, err := p.Run()
	if err != nil {
		log.Println(err.Error())
		return c.readValue(message, validate)
	}
	return r
}

func initProjectConfig() {
	log.Println("Setup Storm: Enter -q to exit")
	reader := &consoleReader{}
	name := reader.readValue("Project Name", func(s string) error {
		if len(s) < 3 {
			return errors.New("invalid project name. Project name length must be more than 3")
		}
		return nil
	})
	if name == "" {
		os.Exit(1)
	}
	serverUrl := reader.readValue("Server Address", func(s string) error {
		_, err := url.Parse(s)
		if err != nil {
			return errors.New("invalid server url")
		}
		return nil
	})
	if serverUrl == "" {
		os.Exit(1)
	}

	serverCode := reader.readValue("Server Authentication Code", func(s string) error {
		return nil
	})
	cfg := &client.Config{
		ServerUrl:      serverUrl,
		ServerAuthCode: serverCode,
		AppName:        name,
	}
	if err := client.Generate(cfg); err != nil {
		log.Println("Failed to init project: ", err)
	}
	log.Println("Project setup completed successfully")
	os.Exit(0)
}

func loadConfig() (*client.Config, string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, wd, err
	}
	path := filepath.Join(wd, "storm_config.json")
	in, err := os.Open(path)
	if err != nil {
		return nil, wd, err
	}
	cfg, err := client.Parse(in)
	if err != nil {
		return nil, wd, err
	}
	return cfg, wd, nil
}

func runClient() {
	cfg, wd, err := loadConfig()
	if err != nil {
		log.Fatal("failed to open configuration: ", err)
	}
	c, err := client.NewStormClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Building Binary...")
	if err := c.BuildBinary(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Println("Starting Deployment...")
	r := &client.DeploymentResult{}
	binPath := fmt.Sprintf("%s/%s", wd, cfg.AppName)
	if err := c.DeployApp(binPath, r); err != nil {
		log.Println("Deployment failed: ", err)
	} else {
		log.Println("App deployed: ", r.Data.AccessUrl)
	}
}

func getLogs() {
	cfg, _, err := loadConfig()
	if err != nil {
		log.Fatal("failed to open configuration: ", err)
	}

	c, err := client.NewStormClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	logs, err := c.GetAppLogs()
	if err != nil {
		log.Println(err)
	}else {
		log.Println(logs)
	}
	os.Exit(0)
}