package main

import (
	"flag"
	jcw "github.com/advantageous/systemd-cloud-watch/cloud-watch"
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
		println("config file name must be set!")
		os.Exit(2)
	}

	config := jcw.CreateConfig(configFilename, logger)
	logger = jcw.NewSimpleLogger("main", config)
	journal := jcw.CreateJournal(config, logger)
	repeater := jcw.CreateRepeater(config, logger)

	jcw.NewRunner(journal, repeater, logger, config)

}

func usage(logger *jcw.Logger) {
	logger.Error.Println("Usage: systemd-cloud-watch  <config-file>")
	flag.PrintDefaults()
}
