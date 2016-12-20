package cloud_watch

import 	lg "github.com/advantageous/go-logback/logging"

func CreateConfig(configFilename string, logger lg.Logger) *Config {

	config, err := LoadConfig(configFilename, logger)
	if err != nil {
		logger.Error("Unable to load config", err, configFilename)
		panic("Unable to create config")
	}
	return config
}

func CreateJournal(config *Config, logger lg.Logger) Journal {

	journal, err := NewJournal(config)
	if err != nil {
		logger.Error("Unable to load journal", err)
		panic("Unable to create journal")
	}
	journal.AddLogFilters(config)
	return journal

}

func CreateRepeater(config *Config, logger lg.Logger) JournalRepeater {

	var repeater JournalRepeater
	var err error

	if !config.MockCloudWatch {
		logger.Info("Creating repeater that is conneting to AWS cloud watch")
		session := NewAWSSession(config)
		repeater, err = NewCloudWatchJournalRepeater(session, nil, config)

	} else {
		logger.Warn("Creating MOCK repeater")
		repeater = NewMockJournalRepeater()
	}

	if err != nil {
		panic("Unable to create repeater " + err.Error())
	}
	return repeater

}
