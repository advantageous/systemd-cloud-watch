package cloud_watch

import (
	"fmt"
	"time"
	"os"
	"os/signal"
	"syscall"
	//	"github.com/coreos/go-systemd/sdjournal"
)

func makeTerminateChannel() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return ch
}

func ReadOneRecord(journal Journal, outChannel chan <- Record, logger *Logger, config *Config,
instanceId string) {

	debug := config.Debug

	count, err := journal.Next()
	if err != nil {

		logger.Error.Printf("error reading from journal: %s", err)

		outChannel <- newErrorRecord(instanceId,
			fmt.Errorf("error reading from journal: %s", err),
		)
		if debug {
			logger.Info.Println("Waiting for two seconds after error")
		}
		time.Sleep(2 * time.Second)
	} else if count <= 0 {
		if debug {
			logger.Info.Println("Waiting for two seconds")
		}
		journal.Wait(2 * time.Second)
	} else if count > 0 {

		if debug {
			logger.Info.Println("No errors, reading log")
		}

		record, err := NewRecord(journal, logger, config)
		record.InstanceId = instanceId

		if err != nil {

			if debug {
				logger.Info.Println("Error reading record", record)
			}

			outChannel <- newErrorRecord(instanceId,
				fmt.Errorf("error unmarshalling record: %s", err),
			)
		} else {

			if debug {
				logger.Info.Println("Read record", record)
			}

			outChannel <- *record
		}
	}

}

func ReadRecords(journal Journal, outChannel chan <- Record, logger *Logger, config *Config) {

	if logger == nil {
		logger = NewSimpleLogger("record reader ", config)
	}

	termC := makeTerminateChannel()
	instanceId := config.EC2InstanceId

	checkTerminate := func() bool {
		select {
		case <-termC:
			close(outChannel)
			return true
		default:
			return false
		}
	}

	for {
		if checkTerminate() {
			logger.Error.Printf("OS signal terminated")
			return
		}
		ReadOneRecord(journal, outChannel, logger, config, instanceId)
	}
}



// BatchRecords consumes a channel of individual records and produces
// a channel of slices of record pointers in sizes up to the given
// batch size.
// If records don't show up fast enough, smaller batches will be returned
// each second as long as at least one item is in the buffer.
func BatchRecords(inputRecordChannel <-chan Record, outputBatchRecords chan <- []Record, logger *Logger, config *Config) {

	if logger == nil {
		logger = NewSimpleLogger("batcher", config)
	}

	batchSize := config.BufferSize

	// We have two buffers here so that we can fill one while the
	// caller is working on the other. The caller is therefore
	// guaranteed that the returned slice will remain valid until
	// the next read of the batches channel.
	var bufs [2][]Record
	bufs[0] = make([]Record, batchSize)
	bufs[1] = make([]Record, batchSize)
	var record Record
	var more bool
	currentBuf := 0
	countOfRecords := 0
	timer := time.NewTimer(time.Second)
	timer.Stop()

	for {
		select {
		case record, more = <-inputRecordChannel:
			if !more {
				close(outputBatchRecords)
				return
			}
			bufs[currentBuf][countOfRecords] = record
			countOfRecords++
			if countOfRecords < batchSize {
				// If we've just added our first record then we'll
				// start the batch timer.
				if countOfRecords == 1 {
					timer.Reset(time.Second)
				}
				// Not enough records yet, so wait again.
				continue
			}
			break
		case <-timer.C:
			break
		}

		timer.Stop()
		if countOfRecords == 0 {
			continue
		}

		// If we manage to fall out here then either the buffer is full
		// or the batch timer expired. Either way it's time for us to
		// emit a batch.
		outputBatchRecords <- bufs[currentBuf][0:countOfRecords]

		// Switch buffers before we start building the next batch.
		currentBuf = (currentBuf + 1) % 2
		countOfRecords = 0
	}
}

func newErrorRecord(instanceId string, err error) Record {
	return Record{
		InstanceId: instanceId,
		Command:  "journald-cloudwatch-logs",
		Priority: ERROR,
		Message:  err.Error(),
	}
}

func CreateConfig(configFilename string, logger *Logger) *Config {

	config, err := LoadConfig(configFilename, logger)
	if err != nil {
		logger.Error.Println("Unable to load config", err, configFilename)
		panic("Unable to create config")
	}
	return config
}

func CreateJournal(config *Config, logger *Logger) Journal {

	journal, err := NewJournal(config)
	if err != nil {
		logger.Error.Println("Unable to load journal", err)
		panic("Unable to create journal")
	}
	journal.AddLogFilters(config)
	return journal

}

func CreateRepeater(config *Config, logger *Logger) JournalRepeater {

	var repeater JournalRepeater
	var err error

	if !config.MockCloudWatch {
		logger.Info.Println("Creating repeater that is conneting to AWS cloud watch")
		session := NewAWSSession(config)
		repeater, err = NewCloudWatchJournalRepeater(session, nil, config)

	} else {
		logger.Warning.Println("Creating MOCK repeater")
		repeater = NewMockJournalRepeater()
	}

	if err != nil {
		panic("Unable to create repeater " + err.Error())
	}
	return repeater

}

func positionCursor(journal Journal, logger *Logger, config *Config) {

	if config.Tail {
		err := journal.SeekTail()
		if err != nil {
			logger.Error.Println("Unable to seek to end of systemd journal", err)
			panic("Unable to seek to end of systemd journal")
		} else {
			logger.Info.Println("Success: Seek to end of systemd journal")
		}

		count, err := journal.PreviousSkip(uint64(config.Rewind))
		if err != nil {
			logger.Error.Println("Unable to rewind after seeking to end of systemd journal", config.Rewind)
			panic("Unable to rewind systemd journal ")
		} else {
			logger.Info.Println("Success: Rewind", config.Rewind, count)
		}
	} else {
		err := journal.SeekHead()
		if err != nil {
			logger.Error.Println("Unable to seek to head of systemd journal", err)
			panic("Unable to seek to end of systemd journal")
		} else {
			logger.Info.Println("Success: Seek to head of systemd journal")
		}

	}

}

func RunWorkers(journal Journal, repeater JournalRepeater, logger *Logger, config *Config) {

	logger.Info.Println("Starting up systemd cloudwatch")

	defer journal.Close()
	defer logger.Info.Println("Leaving RunWorkers method")

	positionCursor(journal, logger, config)

	records := make(chan Record)
	batches := make(chan []Record)

	go ReadRecords(journal, records, nil, config)
	go BatchRecords(records, batches, nil, config)

	for batch := range batches {
		if (config.Debug) {
			logger.Info.Printf("Writing records %d", len(batch))
		}
		err := repeater.WriteBatch(batch)
		if err != nil {
			logger.Error.Panic("Failed to write to cloudwatch:", err)
		}
	}

}