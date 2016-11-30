package main

import (
	jcw  "github.com/RichardHightower/systemd-cloud-watch/cloud-watch"
	"flag"
	"os"
)

var help = flag.Bool("help", false, "set to true to show this help")

func main() {

	logger := jcw.NewSimpleLogger("main", nil)

	flag.Parse()

	if *help {
		usage(logger)
		os.Exit(0)
	}

	configFilename := flag.Arg(0)
	if configFilename == "" {
		usage(logger)
		os.Exit(1)
	}

	err := jcw.RunWorkers(configFilename, logger)
	if err != nil {
		logger.Error.Printf("Error %s", err)
		os.Exit(2)
	}
}

func usage(logger *jcw.Logger) {
	logger.Error.Println("Usage: systemd-cloud-watch  <config-file>")
	flag.PrintDefaults()
}

