package cloud_watch

import "testing"

func TestConfig(t *testing.T) {

	logger := InitSimpleLog("test", nil)

	data := `
log_group="dcos-logstream-test"
state_file="/var/lib/journald-cloudwatch-logs/state-test"
log_priority=3
debug=true
fields=["Foo", "Bar"]
	`
	config, err := LoadConfigFromString(data, logger)

	if err!=nil {
		t.Logf("Unable to parse config %s", err)
		t.Fail()
	}

	if config == nil {
		t.Log("Config is nil")
		t.Fail()
	}

	if len(config.AllowedFields) != 2 {
		t.Log("Fields not read")
		t.Fail()
	}

	logger.Info.Println(config.AllowedFields)


	if config.AllowedFields[0] != "Foo" {
		t.Log("Field Value Foo not present")
		t.Fail()
	}

	if !config.AllowField("Foo") {
		t.Log("Field Value Foo should be allowed")
		t.Fail()
	}

}

func TestLogOmitField(t *testing.T) {

	logger := InitSimpleLog("test", nil)

	data := `omit_fields=["Foo", "Bar"]`
	config, _ := LoadConfigFromString(data, logger)


	if config.AllowField("Foo") {
		t.Log("Field Value Foo should NOT allowed")
		t.Fail()
	}

}