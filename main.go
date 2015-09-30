package main

import (
	"github.com/elastic/libbeat/beat"
)

// You can overwrite these, e.g.: go build -ldflags "-X main.Version 1.0.0-beta3"
var Version = "1.0.0-beta1"
var Name = "dockerbeat"

func main() {

	tb := &Dockerbeat{}

	b := beat.NewBeat(Name, Version, tb)

	b.CommandLineSetup()

	b.LoadConfig()
	tb.Config(b)

	b.Run()

}
