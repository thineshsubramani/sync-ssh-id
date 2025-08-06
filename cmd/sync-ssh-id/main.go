package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/thineshsubramani/sync-ssh-id/internal/cli"
)

func main() {
	_ = godotenv.Load() // load global .env silently

	opts := cli.ParseFlags() // parse args & flags

	switch {
	case opts.Interactive:
		if err := cli.RunInteractive(opts); err != nil {
			log.Fatalf("[interactive] error: %v", err)
		}
	default:
		if err := cli.RunInventory(opts); err != nil {
			log.Fatalf("[inventory] error: %v", err)
		}
	}

}
