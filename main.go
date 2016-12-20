package main

import (
	"flag"
	jcw "github.com/advantageous/systemd-cloud-watch/cloud-watch"
	"os"
	lg "github.com/advantageous/go-logback/logging"
)

var help = flag.Bool("help", false, "set to true to show this help")

func main() {

	logger := lg.NewSimpleLogger("main")

	flag.Parse()

	if *help {
		usage(logger)
		os.Exit(0)
	}

	configFilename := flag.Arg(0)
	if configFilename == "" {
		usage(logger)
		println("config file name must be set!")
		os.Exit(2)
	}

	config := jcw.CreateConfig(configFilename, logger)
	journal := jcw.CreateJournal(config, logger)
	repeater := jcw.CreateRepeater(config, logger)

	jcw.NewRunner(journal, repeater, logger, config)

}

func usage(logger lg.Logger) {
	logger.Error("Usage: systemd-cloud-watch  <config-file>")
	flag.PrintDefaults()
}
