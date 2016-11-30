package cloud_watch

import "time"

type JournalRepeater interface {
	// Close closes a journal opened with NewJournal.
	Close() error;
	WriteBatch(records []Record) error;
}

type Journal interface {
	// Close closes a journal opened with NewJournal.
	Close() error;

	// Next advances the read pointer into the journal by one entry.
	Next() (uint64, error);

	// NextSkip advances the read pointer by multiple entries at once,
	// as specified by the skip parameter.
	NextSkip(skip uint64) (uint64, error);

	// Previous sets the read pointer into the journal back by one entry.
	Previous() (uint64, error);

	// PreviousSkip sets back the read pointer by multiple entries at once,
	// as specified by the skip parameter.
	PreviousSkip(skip uint64) (uint64, error);

	// GetDataValue gets the data object associated with a specific field from the
	// current journal entry, returning only the value of the object.
	GetDataValue(field string) (string, error);


	// GetRealtimeUsec gets the realtime (wallclock) timestamp of the current
	// journal entry.
	GetRealtimeUsec() (uint64, error);

	// GetMonotonicUsec gets the monotonic timestamp of the current journal entry.
	GetMonotonicUsec() (uint64, error);

	// GetCursor gets the cursor of the current journal entry.
	GetCursor() (string, error);


	// SeekHead seeks to the beginning of the journal, i.e. the oldest available
	// entry.
	SeekHead() error;

	// SeekTail may be used to seek to the end of the journal, i.e. the most recent
	// available entry.
	SeekTail() error;

	// SeekCursor seeks to a concrete journal cursor.
	SeekCursor(cursor string) error;

	// Wait will synchronously wait until the journal gets changed. The maximum time
	// this call sleeps may be controlled with the timeout parameter.  If
	// sdjournal.IndefiniteWait is passed as the timeout parameter, Wait will
	// wait indefinitely for a journal change.
	Wait(timeout time.Duration) int;
}