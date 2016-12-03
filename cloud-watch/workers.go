package cloud_watch

import (
	"fmt"
	"time"
	"os"
	"os/signal"
	"syscall"
)

func makeTerminateChannel() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return ch
}

func ReadOneRecord(journal Journal, outChannel chan <- Record, logger *Logger, config *Config,
instanceId string)  {


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

func RunWorkers(configFilename string, logger *Logger) error {

	logger.Info.Println("Starting up systemd cloudwatch")

	config, err := LoadConfig(configFilename, logger); if err != nil {
		logger.Error.Println("Unable to load config %s %s", err, configFilename)
		return fmt.Errorf("error opening config: %s %s", err, configFilename)
	}

	journal, err := NewJournal(config)
	journal.AddLogFilters(config)

	if err != nil {
		logger.Error.Println("Unable to load journal %s", err)
		return fmt.Errorf("error opening journal: %s", err)
	}
	defer journal.Close()
	defer logger.Info.Println("Leaving main run method")

	err = journal.SeekTail(); if err != nil {
		logger.Error.Println("Unable to seek to end of systemd journal")
		return err
	} else {
		logger.Info.Println("Success: Seek to end of systemd journal")
	}

	_, err = journal.PreviousSkip(10); if err != nil {
		logger.Error.Println("Unable to rewind 10 after seeking to end of systemd journal")
		return err
	} else {
		logger.Info.Println("Success: Rewind 10")
	}

	records := make(chan Record)
	batches := make(chan []Record)

	go ReadRecords(journal, records, nil, config)
	go BatchRecords(records, batches, nil, config)

	session := NewAWSSession(config)

	repeater, err := NewCloudWatchJournalRepeater(session, nil, config)

	for batch := range batches {
		if (config.Debug) {
			logger.Info.Printf("Writing records %d", len(batch))
		}
		err := repeater.WriteBatch(batch)
		if err != nil {
			return fmt.Errorf("Failed to write to cloudwatch: %s", err)
		}
	}
	return nil

	//
	//for {
	//
	//	count, err := journal.Next()
	//	if err != nil {
	//		logger.Error.Printf("Unable to read from systemd journal %s", err)
	//		// It's likely that we didn't actually advance here, so
	//		// we should wait a bit so we don't spin the CPU at 100%
	//		// when we run into errors.
	//		time.Sleep(2 * time.Second)
	//		continue
	//	} else if count == 0 {
	//		// If there's nothing new in the stream then we'll
	//		// wait for something new to show up.
	//		// to gracefully terminate because of this. It'd be nicer
	//		// to stop waiting if we get a termination signal, but
	//		// this will do for now.
	//		logger.Error.Println("Systemd journal is empty")
	//		journal.Wait(2 * time.Second)
	//		continue
	//	} else {
	//		value, _ := journal.GetDataValue("MESSAGE")
	//		logger.Debug.Printf("Message %s", value)
	//	}
	//
	//}
	//
	return nil
}