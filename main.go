package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/ingensi/dockerbeat/beater"
)

func main() {
	err := beat.Run("dockerbeat", "", beater.New())
	if err != nil {
		os.Exit(1)
	}
}
