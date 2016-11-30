package cloud_watch



func NewJournal (config *Config) (Journal, error) {
	return &TestJournal{}, nil
}
