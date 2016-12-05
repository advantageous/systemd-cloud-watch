package cloud_watch

import (
	"time"
	"sync/atomic"
)

type MockJournal interface {
	Journal
	SetCount(uint64)
	SetError (error)

}

type TestJournal struct {
	values  map[string]string
	logger  *Logger
	count   int64
	err     error
}

type MockJournalRepeater struct {
	logger  *Logger
}

func (repeater *MockJournalRepeater) Close() error {
	return nil
}

func (repeater *MockJournalRepeater) WriteBatch(records []Record) error {

	for _, record := range records {

		priority := string(PriorityJsonMap[record.Priority])

		switch record.Priority {

		case EMERGENCY:
			repeater.logger.Error.Println(priority, "------", record.Message)
		case ALERT:
			repeater.logger.Error.Println(priority, "------", record.Message)

		case CRITICAL:
			repeater.logger.Error.Println(priority, "------", record.Message)
		case ERROR:
			repeater.logger.Error.Println(priority, "------", record.Message)
		case NOTICE:
			repeater.logger.Warning.Println(priority, "------", record.Message)

		case WARNING:
			repeater.logger.Warning.Println(priority, "------", record.Message)

		case INFO:
			repeater.logger.Info.Println(priority, "------", record.Message)

		case DEBUG:
			repeater.logger.Debug.Println(priority, "------", record.Message)

		default:
			repeater.logger.Debug.Println("?????", priority, "------", record.Message)

		}

	}
	return nil
}

func NewMockJournalRepeater() (repeater *MockJournalRepeater) {
	return &MockJournalRepeater{NewSimpleLogger("mock-repeater", nil)}
}


func (journal *TestJournal) SetCount(count uint64) {

	atomic.StoreInt64(&journal.count, int64(count))


}

func (journal *TestJournal) SetError(err error) {
	journal.err = err

}

func NewJournalWithMap(values map[string]string) Journal {
	logger := NewSimpleLogger("test-journal", nil)
	return &TestJournal{
		values: values,
		logger: logger,
	}
}

func (journal *TestJournal) Close() error {
	journal.logger.Info.Println("Close")
	return nil
}


// Next advances the read pointer into the journal by one entry.
func (journal *TestJournal) Next() (uint64, error) {
	journal.logger.Info.Println("Next")

	var count uint64

	if (journal.count > 0) {
		count = uint64(journal.count)
		atomic.AddInt64(&journal.count, -1)
	} else {
		count = 0
	}



	return uint64(count), nil

}

// NextSkip advances the read pointer by multiple entries at once,
// as specified by the skip parameter.
func (journal *TestJournal) NextSkip(skip uint64) (uint64, error) {
	journal.logger.Info.Println("Next Skip")
	return uint64(journal.count), nil
}

// Previous sets the read pointer into the journal back by one entry.
func (journal *TestJournal) Previous() (uint64, error) {
	journal.logger.Info.Println("Previous")
	return uint64(journal.count), nil
}

// PreviousSkip sets back the read pointer by multiple entries at once,
// as specified by the skip parameter.
func (journal *TestJournal) PreviousSkip(skip uint64) (uint64, error) {
	journal.logger.Info.Println("Previous Skip")
	return uint64(journal.count), nil
}

// GetDataValue gets the data object associated with a specific field from the
// current journal entry, returning only the value of the object.
func (journal *TestJournal) GetDataValue(field string) (string, error) {
	if journal.count < 0 {
		panic("ARGH")
	}
	journal.logger.Info.Println("GetDataValue")
	return journal.values[field], nil
}


// GetRealtimeUsec gets the realtime (wallclock) timestamp of the current
// journal entry.
func (journal *TestJournal) GetRealtimeUsec() (uint64, error) {
	journal.logger.Info.Println("GetRealtimeUsec")
	return 1480549576015541 / 1000, nil
}

func (journal *TestJournal) AddLogFilters(config *Config) {
	journal.logger.Info.Println("AddLogFilters")
}

// GetMonotonicUsec gets the monotonic timestamp of the current journal entry.
func (journal *TestJournal) GetMonotonicUsec() (uint64, error) {
	journal.logger.Info.Println("GetMonotonicUsec")
	return uint64(journal.count), nil
}

// GetCursor gets the cursor of the current journal entry.
func (journal *TestJournal) GetCursor() (string, error) {
	journal.logger.Info.Println("GetCursor")
	return "abc-123", nil
}


// SeekHead seeks to the beginning of the journal, i.e. the oldest available
// entry.
func (journal *TestJournal) SeekHead() error {
	return nil
}

// SeekTail may be used to seek to the end of the journal, i.e. the most recent
// available entry.
func (journal *TestJournal) SeekTail() error {
	return nil
}

// SeekCursor seeks to a concrete journal cursor.
func (journal *TestJournal) SeekCursor(cursor string) error {
	return nil
}

// Wait will synchronously wait until the journal gets changed. The maximum time
// this call sleeps may be controlled with the timeout parameter.  If
// sdjournal.IndefiniteWait is passed as the timeout parameter, Wait will
// wait indefinitely for a journal change.
func (journal *TestJournal) Wait(timeout time.Duration) int {
	return 5
}