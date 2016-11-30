package main

import (
	jcw  "github.com/RichardHightower/systemd-cloud-watch/cloud-watch"
	"fmt"
	"flag"
	"os"
	"time"
)

var help = flag.Bool("help", false, "set to true to show this help")

func main() {

	logger := jcw.InitSimpleLog("main", nil)

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

	err := run(configFilename, logger)
	if err != nil {
		logger.Error.Printf("Error %s", err)
		os.Exit(2)
	}
}

func usage(logger *jcw.Logger) {
	logger.Error.Println("Usage: systemd-cloud-watch  <config-file>")
	flag.PrintDefaults()
}

func run(configFilename string, logger *jcw.Logger) error {

	logger.Info.Println("Starting up systemd cloudwatch")

	config, err := jcw.LoadConfig(configFilename, logger); if err != nil {
		logger.Error.Println("Unable to load config %s %s", err, configFilename)
		return fmt.Errorf("error opening config: %s %s", err, configFilename)
	}

	logger = jcw.InitSimpleLog("main", config)

	journal, err := jcw.NewJournal(config)

	if err != nil {
		logger.Error.Println("Unable to load journal %s", err)
		return fmt.Errorf("error opening journal: %s", err)
	}
	defer journal.Close()
	defer logger.Info.Println("Leaving main run method")

	err = journal.SeekTail(); if err != nil {
		return err
	}

	logger.Info.Println("Went to end of file")

	_, err = journal.PreviousSkip(10); if err != nil {
		return err
	}

	logger.Debug.Println("PreviousSkip success")

	for {

		count, err := journal.Next()
		if err != nil {
			logger.Error.Printf("Unable to read from systemd journal %s", err)
			// It's likely that we didn't actually advance here, so
			// we should wait a bit so we don't spin the CPU at 100%
			// when we run into errors.
			time.Sleep(2 * time.Second)
			continue
		} else if count == 0 {
			// If there's nothing new in the stream then we'll
			// wait for something new to show up.
			// to gracefully terminate because of this. It'd be nicer
			// to stop waiting if we get a termination signal, but
			// this will do for now.
			logger.Error.Println("Systemd journal is empty")
			journal.Wait(2 * time.Second)
			continue
		} else {
			value, _ := journal.GetDataValue("MESSAGE")
			logger.Debug.Printf("Message %s", value)
		}

	}

	return nil
}
