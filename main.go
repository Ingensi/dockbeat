package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/ingensi/dockbeat/beater"
)

func main() {
	err := beat.Run("dockbeat", "1.0.0", beater.New())
	if err != nil {
		os.Exit(1)
	}
}
