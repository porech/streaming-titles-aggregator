package main

import (
	"log"
	"os"

	"github.com/porech/streaming-titles-aggregator/internal/server"
	_ "github.com/porech/streaming-titles-aggregator/internal/source"
)

func main() {
	configPath := "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	if err := server.Run(configPath); err != nil {
		log.Fatal(err)
	}
}
