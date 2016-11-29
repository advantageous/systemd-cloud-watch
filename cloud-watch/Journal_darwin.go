package cloud_watch

import "time"

type TestJournal struct {
	values map[string]string
}

func NewJournal () (Journal, error) {
	return &TestJournal{}, nil
}

func NewJournaWithMap (values map[string]string) Journal {
	return &TestJournal{values}
}


func (journal *TestJournal) Close() error {
	return nil
}


// Next advances the read pointer into the journal by one entry.
func (journal *TestJournal) Next() (uint64, error) {
	return 1, nil
}

// NextSkip advances the read pointer by multiple entries at once,
// as specified by the skip parameter.
func (journal *TestJournal) NextSkip(skip uint64) (uint64, error){
	return 1, nil
}

// Previous sets the read pointer into the journal back by one entry.
func (journal *TestJournal) Previous() (uint64, error){
	return 1, nil
}

// PreviousSkip sets back the read pointer by multiple entries at once,
// as specified by the skip parameter.
func (journal *TestJournal) PreviousSkip(skip uint64) (uint64, error){
	return 1, nil
}

// GetDataValue gets the data object associated with a specific field from the
// current journal entry, returning only the value of the object.
func (journal *TestJournal) GetDataValue(field string) (string, error){
	return journal.values[field], nil
}


// GetRealtimeUsec gets the realtime (wallclock) timestamp of the current
// journal entry.
func (journal *TestJournal) GetRealtimeUsec() (uint64, error){
	return 1, nil
}

// GetMonotonicUsec gets the monotonic timestamp of the current journal entry.
func (journal *TestJournal) GetMonotonicUsec() (uint64, error){
	return 1, nil
}

// GetCursor gets the cursor of the current journal entry.
func (journal *TestJournal) GetCursor() (string, error){
	return "abc-123", nil
}


// SeekHead seeks to the beginning of the journal, i.e. the oldest available
// entry.
func (journal *TestJournal) SeekHead() error{
	return nil
}

// SeekTail may be used to seek to the end of the journal, i.e. the most recent
// available entry.
func (journal *TestJournal) SeekTail() error{
	return nil
}

// SeekCursor seeks to a concrete journal cursor.
func (journal *TestJournal) SeekCursor(cursor string) error{
	return nil
}

// Wait will synchronously wait until the journal gets changed. The maximum time
// this call sleeps may be controlled with the timeout parameter.  If
// sdjournal.IndefiniteWait is passed as the timeout parameter, Wait will
// wait indefinitely for a journal change.
func (journal *TestJournal) Wait(timeout time.Duration) int{
	return 5
}