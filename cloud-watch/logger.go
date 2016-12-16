package cloud_watch

import (
	"io"
	"log"
	"os"
	"io/ioutil"
)

type Logger struct {
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

func NewSimpleLogger(name string, config *Config) *Logger {

	if (config == nil) {
		return NewLogger(name, ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	} else if config.Debug {
		return NewLogger(name, os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		return NewLogger(name, ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	}

}

func NewLogger(name string, traceHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) *Logger {

	logger := Logger{
	}

	logger.Debug = log.New(traceHandle,
		name + " DEBUG: ",
		log.Ldate | log.Ltime | log.Lshortfile)

	logger.Info = log.New(infoHandle,
		name + " INFO: ",
		log.Ldate | log.Ltime | log.Lshortfile)

	logger.Warning = log.New(warningHandle,
		name + " WARNING: ",
		log.Ldate | log.Ltime | log.Lshortfile)

	logger.Error = log.New(errorHandle,
		name + " ERROR: ",
		log.Ldate | log.Ltime | log.Lshortfile)

	return &logger

}
