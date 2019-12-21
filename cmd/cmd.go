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
	args := os.Args
	if len(args) > 1 {
		arg := args[1]
		if arg == "init" {
			if err := storm.InitDefaultConfig(configPath); err != nil {
				log.Println(err)
				os.Exit(1)
			} else {
				log.Printf("Created default config: %s", configPath)
				os.Exit(0)
			}
		}
	}

	svr, err := storm.NewServer(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := svr.Run(":9870"); err != nil {
		log.Fatal(err)
	}
}
