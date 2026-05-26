package main

import (
	"flag"
	"log"

	"github.com/porech/streaming-titles-aggregator/internal/server"
	_ "github.com/porech/streaming-titles-aggregator/internal/source"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	configPath := flag.String("config", "streams.json", "path to configuration file")
	flag.Parse()

	if err := server.Run(*addr, *configPath); err != nil {
		log.Fatal(err)
	}
}
