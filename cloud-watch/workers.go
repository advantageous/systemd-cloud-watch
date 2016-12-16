package cloud_watch

import (
	"fmt"
	q "github.com/advantageous/go-qbit/qbit"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Runner struct {
	records         []*Record
	bufferSize      int
	logger          *Logger
	journalRepeater JournalRepeater
	journal         Journal
	batchCounter    uint64
	idleCounter     uint64
	emptyCounter    uint64
	lastMetricTime  int64
	queueManager    q.QueueManager
	config          *Config
	debug           bool
	instanceId      string
}

func (r *Runner) Stop() {
	r.queueManager.Stop()
}
func (r *Runner) addToCloudWatchBatch(record *Record) {

	r.records = append(r.records, record)

	if len(r.records) >= r.bufferSize {
		r.sendBatch()
	}
}

func (r *Runner) sendBatch() {

	if len(r.records) > 0 {
		batchToSend := r.records
		r.records = make([]*Record, 0)
		err := r.journalRepeater.WriteBatch(batchToSend)
		if err != nil {
			r.logger.Error.Println("Failed to write to cloudwatch batch size = : %d %s %v",
				len(r.records), err.Error(), err)
		}

	}
}

func NewRunnerInternal(journal Journal, repeater JournalRepeater, logger *Logger, config *Config, start bool) *Runner {

	if repeater == nil {
		panic("Repeater can't be nil")
	}
	r := &Runner{journal: journal,
		journalRepeater: repeater,
		logger:          logger,
		config:          config,
		debug:           config.Debug,
		instanceId:      config.EC2InstanceId,
		bufferSize:      config.CloudWatchBufferSize}

	if logger == nil {
		logger = NewSimpleLogger("record reader ", config)
	}

	r.queueManager = q.NewQueueManager(config.QueueChannelSize,
		config.QueueBatchSize,
		time.Duration(config.QueuePollDurationMS)*time.Millisecond,
		q.NewQueueListener(&q.QueueListener{

			ReceiveFunc: func(item interface{}) {
				r.addToCloudWatchBatch(item.(*Record))
			},
			EndBatchFunc: func() {
				r.sendBatch()
				r.batchCounter++
			},
			IdleFunc: func() {
				r.sendBatch()
				now := time.Now().Unix()
				if now-r.lastMetricTime > 120 {
					now = r.lastMetricTime
					r.logger.Info.Printf("Systemd CloudWatch: batches sent %d, idleCount %d,  emptyCount %d",
						r.batchCounter, r.idleCounter, r.emptyCounter)
				}
				r.idleCounter++
			},
			EmptyFunc: func() {
				r.sendBatch()
				r.emptyCounter++
			},
		}))

	r.lastMetricTime = time.Now().Unix()
	r.positionCursor()

	if start {
		signalChannel := r.makeTerminateChannel()

		go func() {
			<-signalChannel
			r.queueManager.Stop()
		}()

		r.readRecords()
	}

	return r
}
func NewRunner(journal Journal, repeater JournalRepeater, logger *Logger, config *Config) *Runner {
	return NewRunnerInternal(journal, repeater, logger, config, true)

}

func (r *Runner) makeTerminateChannel() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return ch
}

func (r *Runner) readOneRecord() (*Record, bool, error) {

	count, err := r.journal.Next()
	if err != nil {
		return nil, false, err
	} else if count > 0 {
		if r.debug {
			r.logger.Info.Println("No errors, reading log")
		}
		record, err := NewRecord(r.journal, r.logger, r.config)
		record.InstanceId = r.instanceId
		if err != nil {
			return nil, false, fmt.Errorf("error unmarshalling record: %v", err)
		}
		if r.debug {
			r.logger.Info.Println("Read record", record)
		}
		return record, true, nil
	} else {

		if r.debug {
			r.logger.Info.Println("Waiting for two seconds")
		}
		r.journal.Wait(2 * time.Second)
		return nil, false, nil
	}

}

func (r *Runner) readRecords() {

	sendQueue := r.queueManager.SendQueueWithAutoFlush(time.Duration(r.config.FlushLogEntries) * time.Millisecond)

	for {

		record, isReadRecord, err := r.readOneRecord()

		if err == nil && isReadRecord && record != nil {
			sendQueue.Send(record)
		}

		if err != nil {
			r.logger.Error.Println("Error reading record", err)
		}

		if !isReadRecord {
			if r.queueManager.Stopped() {
				r.logger.Info.Println("Got stop message")
				break
			}
		}

	}

}

func (r *Runner) positionCursor() {

	if r.config.Tail {
		err := r.journal.SeekTail()
		if err != nil {
			r.logger.Error.Println("Unable to seek to end of systemd journal", err)
			panic("Unable to seek to end of systemd journal")
		} else {
			r.logger.Info.Println("Success: Seek to end of systemd journal")
		}

		count, err := r.journal.PreviousSkip(uint64(r.config.Rewind))
		if err != nil {
			r.logger.Error.Println("Unable to rewind after seeking to end of systemd journal", r.config.Rewind)
			panic("Unable to rewind systemd journal ")
		} else {
			r.logger.Info.Println("Success: Rewind", r.config.Rewind, count)
		}
	} else {
		err := r.journal.SeekHead()
		if err != nil {
			r.logger.Error.Println("Unable to seek to head of systemd journal", err)
			panic("Unable to seek to end of systemd journal")
		} else {
			r.logger.Info.Println("Success: Seek to head of systemd journal")
		}

	}

}
