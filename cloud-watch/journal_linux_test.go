package cloud_watch

import "testing"

func TestNewJournal(t *testing.T) {

	j, e := NewJournal(nil)

	if e != nil {
		t.Fail()
	}

	if j == nil {
		t.Fail()
	}

	e = j.Close()

	if e != nil {
		t.Fail()
	}

}

func TestSdJournal_Operations(t *testing.T) {
	j, e := NewJournal(nil)

	if e != nil {
		t.Fail()
	}

	j.SeekHead()
	j.Next()

	value, e := j.GetDataValue("MESSAGE")

	if len(value) == 0 {
		t.Logf("Failed value=%s err=%s", value, e)
		t.Fail()
	} else {
		t.Logf("Read value=%s", value)
	}

}
