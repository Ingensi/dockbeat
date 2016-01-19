package main

import (
	dockerbeat "github.com/ingensi/dockerbeat/beat"

	"github.com/elastic/beats/libbeat/beat"
)

// You can overwrite these, e.g.: go build -ldflags "-X main.Version 1.0.0-beta3"
var Version = "1.0.0-beta2"
var Name = "dockerbeat"

func main() {
	beat.Run(Name, Version, dockerbeat.New())
}
