package cloud_watch

import (
	"sync/atomic"
	"time"
	lg "github.com/advantageous/go-logback/logging"
)

type MockJournal interface {
	Journal
	SetCount(uint64)
	SetError(error)
}

type TestJournal struct {
	values map[string]string
	logger lg.Logger
	count  int64
	err    error
}

type MockJournalRepeater struct {
	logger lg.Logger
}

func (repeater *MockJournalRepeater) Close() error {
	return nil
}

func (repeater *MockJournalRepeater) WriteBatch(records []*Record) error {

	for _, record := range records {

		priority := string(PriorityJsonMap[record.Priority])

		switch record.Priority {

		case EMERGENCY:
			repeater.logger.Error(priority, "------", record.Message)
		case ALERT:
			repeater.logger.Error(priority, "------", record.Message)

		case CRITICAL:
			repeater.logger.Error(priority, "------", record.Message)
		case ERROR:
			repeater.logger.Error(priority, "------", record.Message)
		case NOTICE:
			repeater.logger.Warn(priority, "------", record.Message)

		case WARNING:
			repeater.logger.Warn(priority, "------", record.Message)

		case INFO:
			repeater.logger.Info(priority, "------", record.Message)

		case DEBUG:
			repeater.logger.Debug(priority, "------", record.Message)

		default:
			repeater.logger.Debug("?????", priority, "------", record.Message)

		}

	}
	return nil
}

func NewMockJournalRepeater() (repeater *MockJournalRepeater) {
	return &MockJournalRepeater{lg.NewSimpleLogger("mock-repeater")}
}

func (journal *TestJournal) SetCount(count uint64) {

	atomic.StoreInt64(&journal.count, int64(count))

}

func (journal *TestJournal) SetError(err error) {
	journal.err = err

}

func NewJournalWithMap(values map[string]string) Journal {
	logger := lg.NewSimpleLogger("test-journal")
	return &TestJournal{
		values: values,
		logger: logger,
		count:  113,
	}
}

func (journal *TestJournal) Close() error {
	journal.logger.Info("Close")
	return nil
}

// Next advances the read pointer into the journal by one entry.
func (journal *TestJournal) Next() (uint64, error) {
	journal.logger.Debug("Next")

	var count = atomic.LoadInt64(&journal.count)

	if count > 0 {
		atomic.AddInt64(&journal.count, -1)
		return uint64(1), nil
	} else {
		return uint64(0), nil
	}

}

// NextSkip advances the read pointer by multiple entries at once,
// as specified by the skip parameter.
func (journal *TestJournal) NextSkip(skip uint64) (uint64, error) {
	journal.logger.Info("Next Skip")
	return uint64(journal.count), nil
}

// Previous sets the read pointer into the journal back by one entry.
func (journal *TestJournal) Previous() (uint64, error) {
	journal.logger.Info("Previous")
	return uint64(journal.count), nil
}

// PreviousSkip sets back the read pointer by multiple entries at once,
// as specified by the skip parameter.
func (journal *TestJournal) PreviousSkip(skip uint64) (uint64, error) {
	journal.logger.Info("Previous Skip")
	return uint64(journal.count), nil
}

// GetDataValue gets the data object associated with a specific field from the
// current journal entry, returning only the value of the object.
func (journal *TestJournal) GetDataValue(field string) (string, error) {
	if journal.count < 0 {
		panic("ARGH")
	}
	journal.logger.Debug("GetDataValue")
	return journal.values[field], nil
}

// GetRealtimeUsec gets the realtime (wallclock) timestamp of the current
// journal entry.
func (journal *TestJournal) GetRealtimeUsec() (uint64, error) {
	journal.logger.Info("GetRealtimeUsec")
	return 1480549576015541 / 1000, nil
}

func (journal *TestJournal) AddLogFilters(config *Config) {
	journal.logger.Info("AddLogFilters")
}

// GetMonotonicUsec gets the monotonic timestamp of the current journal entry.
func (journal *TestJournal) GetMonotonicUsec() (uint64, error) {
	journal.logger.Info("GetMonotonicUsec")
	return uint64(journal.count), nil
}

// GetCursor gets the cursor of the current journal entry.
func (journal *TestJournal) GetCursor() (string, error) {
	journal.logger.Info("GetCursor")
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
