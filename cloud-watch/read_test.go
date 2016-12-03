package cloud_watch

import (
	"testing"
	"time"
	"errors"
)


var readTestMap = map[string]string{
	"__CURSOR": "s=6c072e0567ff423fa9cb39f136066299;i=3;b=923def0648b1422aa28a8846072481f2;m=65ee792c;t=542783a1cc4e0;x=7d96bf9e60a6512b",
	"__REALTIME_TIMESTAMP": "1480459022025952",
	"__MONOTONIC_TIMESTAMP": "1710127404",
	"_BOOT_ID": "923def0648b1422aa28a8846072481f2",
	"PRIORITY": "6",
	"_TRANSPORT": "driver",
	"_PID": "712",
	"_UID": "0",
	"_GID": "0",
	"_COMM": "systemd-journal",
	"_EXE": "/usr/lib/systemd/systemd-journald",
	"_CMDLINE": "/usr/lib/systemd/systemd-journald",
	"_CAP_EFFECTIVE": "a80425fb",
	"_SYSTEMD_CGROUP": "c",
	"_MACHINE_ID": "5125015c46bb4bf6a686b5e692492075",
	"_HOSTNAME": "f5076731cfdb",
	"MESSAGE": "Journal started",
	"MESSAGE_ID": "f77379a8490b408bbe5f6940505a777b",
}

const readTestConfigData =  `
log_group="dcos-logstream-test"
state_file="/var/lib/journald-cloudwatch-logs/state-test"
log_priority=3
debug=true
	`



func TestReadFromJournalSuccess(t *testing.T) {

	logger := NewSimpleLogger("read-config-test", nil)
	var journal MockJournal
	journal = NewJournalWithMap(readTestMap).(MockJournal)

	config, _ := LoadConfigFromString(readTestConfigData, logger)
	records := make(chan Record)


	journal.SetCount(1)


	go ReadOneRecord(journal, records, logger, config, "foo-bar")

	var record Record
	var more bool

 	record, more = <-records


	if !more  {
		t.Fail()
	}

	if record == (Record{}) {
		t.Fail()
	}

}



func TestReadFromJournalTimeout(t *testing.T) {

	logger := NewSimpleLogger("read-config-test", nil)
	var journal MockJournal
	journal = NewJournalWithMap(readTestMap).(MockJournal)

	config, _ := LoadConfigFromString(readTestConfigData, logger)
	inputRecordChannel := make(chan Record)

	journal.SetCount(0)
	go ReadOneRecord(journal, inputRecordChannel, logger, config, "foo-bar")

	var record Record
	var more bool


	timer := time.NewTimer(time.Millisecond * 50)

	select {
	case record, more = <-inputRecordChannel:
		if !more {
			return
		}

		if record != (Record{}) {
			t.Fail()
		}
	case <-timer.C:

		logger.Info.Println("Timed out like expeccted")
	}


}



func TestReadFromJournalError(t *testing.T) {

	logger := NewSimpleLogger("read-config-test", nil)
	var journal MockJournal
	journal = NewJournalWithMap(readTestMap).(MockJournal)

	config, _ := LoadConfigFromString(readTestConfigData, logger)
	inputRecordChannel := make(chan Record)

	journal.SetError(errors.New("TEST ERROR"))
	journal.SetCount(1)

	go ReadOneRecord(journal, inputRecordChannel, logger, config, "foo-bar")

	var record Record
	var more bool



	timer := time.NewTimer(time.Millisecond * 50)

	select {
	case record, more = <-inputRecordChannel:
		if !more {
			return
		}

		if record == (Record{}) {
			t.Fatal()
		}
	case <-timer.C:
		t.Fatal()
	}


}



func TestReadAllFromJournal(t *testing.T) {

	logger := NewSimpleLogger("read-config-test", nil)
	var journal MockJournal
	journal = NewJournalWithMap(readTestMap).(MockJournal)

	config, _ := LoadConfigFromString(readTestConfigData, logger)
	inputRecordChannel := make(chan Record)

	journal.SetError(errors.New("TEST ERROR"))

	journal.SetCount(10)

	go ReadRecords(journal, inputRecordChannel, logger, config)

	for index:=0; index < 10; index++ {
		logger.Info.Println("Index", index)

		var record Record

		timer := time.NewTimer(time.Millisecond * 1000)

		select {
		case record = <-inputRecordChannel:

			if record == (Record{}) {
				logger.Info.Println("FAIL")
			}
			timer.Stop()
		case <-timer.C:
			t.Fatal("Timeout")

		}

	}


	timer := time.NewTimer(time.Millisecond * 50)


	record := Record{}

	select {
	case record = <-inputRecordChannel:
		if record != (Record{}) {
			t.Fatal("RECORD NOT EMPTY", record.SeqId)
		}
	case <-timer.C:

		logger.Info.Println("Timed out like expeccted")
	}


}