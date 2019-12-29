package main

import (
	"fmt"
	"github.com/adigunhammedolalekan/storm/client"
	"log"
	"os"
)

func main() {
	cfg := &client.Config{
		ServerUrl:      "http://localhost:9870",
		ServerAuthCode: "NmRiOWI2MDctZGU4NS00ZmQyLThiZTItMGE3MjRmMjkwMjVj",
		AppName:        "testStormApp",
	}
	c, err := client.NewStormClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Building Binary...")
	if err := c.BuildBinary(); err != nil {
		log.Println(err)
	}
	log.Println("Starting Deployment...")
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	r := &client.DeploymentResult{}
	if err := c.DeployApp(fmt.Sprintf("%s/%s", wd, cfg.AppName), r); err != nil {
		log.Println("Deployment failed: ", err)
	} else {
		log.Println("App deployed: ", r.Data.AccessUrl)
	}
}
