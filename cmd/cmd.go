package main

import (
	"github.com/adigunhammedolalekan/storm"
	"log"
	"os"
	"path/filepath"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	configPath := filepath.Join(wd, "storm_config.json")
	svr, err := storm.NewServer(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := svr.Run(":9870"); err != nil {
		log.Fatal(err)
	}
}
