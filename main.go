package main

import (
	dockerbeat "github.com/ingensi/dockerbeat/beat"
	"os"

	"github.com/elastic/libbeat/beat"
	"github.com/elastic/libbeat/logp"
)

// You can overwrite these, e.g.: go build -ldflags "-X main.Version 1.0.0-beta3"
var Version = "1.0.0-beta1"
var Name = "dockerbeat"

func main() {

	d := dockerbeat.New()

	b := beat.NewBeat(Name, Version, d)

	b.CommandLineSetup()

	b.LoadConfig()
	err := d.Config(b)
	if err != nil {
		logp.Critical("Config error: %v", err)
		os.Exit(1)
	}

	b.Run()

}
