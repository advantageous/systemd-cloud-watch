package cloud_watch

import (
	"encoding/json"
	"testing"
)

var testMap = map[string]string{
	"__CURSOR":              "s=6c072e0567ff423fa9cb39f136066299;i=3;b=923def0648b1422aa28a8846072481f2;m=65ee792c;t=542783a1cc4e0;x=7d96bf9e60a6512b",
	"__REALTIME_TIMESTAMP":  "1480459022025952",
	"__MONOTONIC_TIMESTAMP": "1710127404",
	"_BOOT_ID":              "923def0648b1422aa28a8846072481f2",
	"PRIORITY":              "6",
	"_TRANSPORT":            "driver",
	"_PID":                  "712",
	"_UID":                  "0",
	"_GID":                  "0",
	"_COMM":                 "systemd-journal",
	"_EXE":                  "/usr/lib/systemd/systemd-journald",
	"_CMDLINE":              "/usr/lib/systemd/systemd-journald",
	"_CAP_EFFECTIVE":        "a80425fb",
	"_SYSTEMD_CGROUP":       "c",
	"_MACHINE_ID":           "5125015c46bb4bf6a686b5e692492075",
	"_HOSTNAME":             "f5076731cfdb",
	"MESSAGE":               "Journal started",
	"MESSAGE_ID":            "f77379a8490b408bbe5f6940505a777b",
	"SYSLOG_FACILITY":       "5",
}

func TestNewRecord(t *testing.T) {

	journal := NewJournalWithMap(testMap)
	logger := NewSimpleLogger("test", nil)
	data := `
log_group="dcos-logstream-test"
state_file="/var/lib/journald-cloudwatch-logs/state-test"
log_priority=3
debug=true
	`
	config, err := LoadConfigFromString(data, logger)

	record, err := NewRecord(journal, logger, config)

	if err != nil {
		t.Logf("Failed err=%s", err)
		t.Fail()
	}

	if record == nil {
		t.Log("Record nil")
		t.Fail()
	}

	if record.CommandLine != "/usr/lib/systemd/systemd-journald" {
		t.Log("Unable to read cmd line")
		t.Fail()
	}

	if record.TimeUsec != 1480459022025952/1000 {
		t.Logf("Unable to read time stamp %d", record.TimeUsec)
		t.Fail()
	}

}

func TestNewRecordJson(t *testing.T) {

	journal := NewJournalWithMap(testMap)
	logger := NewSimpleLogger("test", nil)
	data := `
log_group="dcos-logstream-test"
state_file="/var/lib/journald-cloudwatch-logs/state-test"
log_priority=3
debug=true
	`
	config, err := LoadConfigFromString(data, logger)

	record, err := NewRecord(journal, logger, config)

	if err != nil {
		t.Logf("Failed err=%s", err)
		t.Fail()
	}

	if record == nil {
		t.Log("Record nil")
		t.Fail()
	}

	if record.CommandLine != "/usr/lib/systemd/systemd-journald" {
		t.Log("Unable to read cmd line")
		t.Fail()
	}

	if record.TimeUsec != 1480459022025952/1000 {
		t.Logf("Unable to read time stamp %d", record.TimeUsec)
		t.Fail()
	}

	jsonDataBytes, err := json.MarshalIndent(record, "", "  ")
	jsonData := string(jsonDataBytes)

	t.Logf(jsonData)

}

func TestLimitFields(t *testing.T) {

	journal := NewJournalWithMap(testMap)
	logger := NewSimpleLogger("test", nil)
	data := `
log_group="dcos-logstream-test"
state_file="/var/lib/journald-cloudwatch-logs/state-test"
log_priority=3
debug=true
fields=["__REALTIME_TIMESTAMP"]

	`
	config, err := LoadConfigFromString(data, logger)

	record, err := NewRecord(journal, logger, config)

	if err != nil {
		t.Logf("Failed err=%s", err)
		t.Fail()
	}

	if record == nil {
		t.Log("Record nil")
		t.Fail()
	}

	if record.CommandLine != "" {
		t.Log("Unable to limit cmd line")
		t.Fail()
	}

	if record.TimeUsec != 1480459022025952/1000 {
		t.Logf("Unable to read time stamp %d", record.TimeUsec)
		t.Fail()
	}

}

func TestOmitFields(t *testing.T) {

	journal := NewJournalWithMap(testMap)
	logger := NewSimpleLogger("test", nil)
	data := `
log_group="dcos-logstream-test"
state_file="/var/lib/journald-cloudwatch-logs/state-test"
log_priority=3
debug=true
omit_fields=["_CMDLINE"]

	`
	config, err := LoadConfigFromString(data, logger)

	record, err := NewRecord(journal, logger, config)

	if err != nil {
		t.Logf("Failed err=%s", err)
		t.Fail()
	}

	if record == nil {
		t.Log("Record nil")
		t.Fail()
	}

	if record.CommandLine != "" {
		t.Log("Unable to limit cmd line")
		t.Fail()
	}

	if record.TimeUsec != 1480459022025952/1000 {
		t.Logf("Unable to read time stamp %d", record.TimeUsec)
		t.Fail()
	}

}
