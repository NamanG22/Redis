package main

import (
	"flag"
	"log"

	"github.com/NamanG22/Redis/config"
	"github.com/NamanG22/Redis/server"
)

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the dice server")
	flag.IntVar(&config.Port, "port", 7379, "port for the dice server")
	flag.Parse()
}

func main() {
	setupFlags()
	log.Println("rolling the dice")
	server.RunSyncTCPServer()
}
