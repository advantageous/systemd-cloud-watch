package cloud_watch

import (
	"github.com/coreos/go-systemd/sdjournal"
	"strconv"
	"time"
)

type SdJournal struct {
	journal *sdjournal.Journal
	logger  *Logger
	debug   bool
}

func NewJournal(config *Config) (Journal, error) {

	logger := NewSimpleLogger("journal", config)

	var debug bool

	if config == nil {
		debug = true
	} else {
		debug = config.Debug
	}

	if config == nil || config.JournalDir == "" {
		journal, err := sdjournal.NewJournal()
		return &SdJournal{
			journal, logger, debug,
		}, err
	} else {
		logger.Info.Printf("using journal dir: %s", config.JournalDir)
		journal, err := sdjournal.NewJournalFromDir(config.JournalDir)

		return &SdJournal{
			journal, logger, debug,
		}, err
	}

}

func (journal *SdJournal) AddLogFilters(config *Config) {

	// Add Priority Filters
	if config.GetJournalDLogPriority() < DEBUG {
		for p, _ := range PriorityJsonMap {
			if p <= config.GetJournalDLogPriority() {
				journal.journal.AddMatch("PRIORITY=" + strconv.Itoa(int(p)))
			}
		}
		journal.journal.AddDisjunction()
	}
}

func (journal *SdJournal) Close() error {
	return journal.journal.Close()
}

// Next advances the read pointer into the journal by one entry.
func (journal *SdJournal) Next() (uint64, error) {
	loc, err := journal.journal.Next()
	if journal.debug {
		journal.logger.Info.Printf("NEXT location %d %v", loc, err)
	}

	return loc, err
}

// NextSkip advances the read pointer by multiple entries at once,
// as specified by the skip parameter.
func (journal *SdJournal) NextSkip(skip uint64) (uint64, error) {
	return journal.journal.NextSkip(skip)
}

// Previous sets the read pointer into the journal back by one entry.
func (journal *SdJournal) Previous() (uint64, error) {
	return journal.journal.Previous()
}

// PreviousSkip sets back the read pointer by multiple entries at once,
// as specified by the skip parameter.
func (journal *SdJournal) PreviousSkip(skip uint64) (uint64, error) {
	return journal.journal.PreviousSkip(skip)
}

// GetDataValue gets the data object associated with a specific field from the
// current journal entry, returning only the value of the object.
func (journal *SdJournal) GetDataValue(field string) (string, error) {
	return journal.journal.GetDataValue(field)
}

// GetRealtimeUsec gets the realtime (wallclock) timestamp of the current
// journal entry.
func (journal *SdJournal) GetRealtimeUsec() (uint64, error) {
	return journal.journal.GetRealtimeUsec()
}

// GetMonotonicUsec gets the monotonic timestamp of the current journal entry.
func (journal *SdJournal) GetMonotonicUsec() (uint64, error) {
	return journal.journal.GetMonotonicUsec()
}

// GetCursor gets the cursor of the current journal entry.
func (journal *SdJournal) GetCursor() (string, error) {
	return journal.journal.GetCursor()
}

// SeekHead seeks to the beginning of the journal, i.e. the oldest available
// entry.
func (journal *SdJournal) SeekHead() error {
	return journal.journal.SeekHead()
}

// SeekTail may be used to seek to the end of the journal, i.e. the most recent
// available entry.
func (journal *SdJournal) SeekTail() error {
	return journal.journal.SeekTail()
}

// SeekCursor seeks to a concrete journal cursor.
func (journal *SdJournal) SeekCursor(cursor string) error {
	return journal.journal.SeekCursor(cursor)
}

// Wait will synchronously wait until the journal gets changed. The maximum time
// this call sleeps may be controlled with the timeout parameter.  If
// sdjournal.IndefiniteWait is passed as the timeout parameter, Wait will
// wait indefinitely for a journal change.
func (journal *SdJournal) Wait(timeout time.Duration) int {
	return journal.journal.Wait(timeout)
}
