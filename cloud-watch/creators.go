package cloud_watch

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
