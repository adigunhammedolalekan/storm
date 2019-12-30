package main

import (
	"bufio"
	"fmt"
	"github.com/adigunhammedolalekan/storm/client"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(wd, "storm_config.json")
	in, err := os.Open(path)
	if err != nil {
		log.Fatal("failed to open configuration: ", err)
	}
	cfg, err := client.Parse(in)
	if err != nil {
		log.Fatal("failed to read configuration: ", err)
	}

	binPath := fmt.Sprintf("%s/%s", wd, cfg.AppName)
	args := os.Args
	if len(args) > 1 {
		arg := args[1]
		switch arg {
		case "init":
			initProjectConfig()
		case "logs":
			getLogs(cfg)
		default:
			runClient(binPath, cfg)
		}
	}
}

type consoleReader struct {
}

func (c *consoleReader) readValue(message string) string {
	buf := bufio.NewReader(os.Stdin)
	log.Print(message)
	for {
		value, err := buf.ReadString(byte('\n'))
		if err != nil {
			log.Println("Try again: ")
			continue
		}
		value = strings.Trim(value, "\n")
		if value == "-q" {
			return ""
		}
		if value != "" {
			return value
		}
	}
}

func initProjectConfig() {
	log.Println("Setup Storm: Enter -q to exit")
	reader := &consoleReader{}
	name := reader.readValue("Project Name: ")
	if name == "" {
		os.Exit(1)
	}
	serverUrl := reader.readValue("Server Address: ")
	if serverUrl == "" {
		os.Exit(1)
	}
	_, err := url.Parse(strings.Trim(serverUrl, "\n"))
	if err != nil {
		log.Println("Invalid server url: ", err)
		os.Exit(1)
	}
	serverCode := reader.readValue("Server Authentication Code: ")

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

func runClient(binPath string, cfg *client.Config) {
	c, err := client.NewStormClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Building Binary...")
	if err := c.BuildBinary(); err != nil {
		log.Println(err)
	}
	log.Println("Starting Deployment...")
	r := &client.DeploymentResult{}
	if err := c.DeployApp(binPath, r); err != nil {
		log.Println("Deployment failed: ", err)
	} else {
		log.Println("App deployed: ", r.Data.AccessUrl)
	}
}

func getLogs(cfg *client.Config) {
	c, err := client.NewStormClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	logs, err := c.GetAppLogs()
	if err != nil {
		log.Println(err)
	}else {
		log.Println()
		log.Println(logs)
	}
	os.Exit(0)
}