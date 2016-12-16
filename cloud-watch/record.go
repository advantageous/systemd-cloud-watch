package cloud_watch

import (
	"reflect"
	"strconv"
	"time"
	//	"encoding/json"
)

type Priority int

var (
	EMERGENCY Priority = 0
	ALERT Priority = 1
	CRITICAL Priority = 2
	ERROR Priority = 3
	WARNING Priority = 4
	NOTICE Priority = 5
	INFO Priority = 6
	DEBUG Priority = 7
)

var PriorityJsonMap = map[Priority][]byte{
	EMERGENCY: []byte("\"EMERG\""),
	ALERT:     []byte("\"ALERT\""),
	CRITICAL:  []byte("\"CRITICAL\""),
	ERROR:     []byte("\"ERROR\""),
	WARNING:   []byte("\"WARNING\""),
	NOTICE:    []byte("\"NOTICE\""),
	INFO:      []byte("\"INFO\""),
	DEBUG:     []byte("\"DEBUG\""),
}

type Record struct {
	InstanceId  string        `json:"instanceId,omitempty"`
	TimeUsec    int64         `json:"-" journald:"__REALTIME_TIMESTAMP"`
	PID         int           `json:"pid,omitempty" journald:"_PID"`
	UID         int           `json:"uid,omitempty" journald:"_UID"`
	GID         int           `json:"gid,omitempty" journald:"_GID"`
	Command     string        `json:"cmdName,omitempty" journald:"_COMM"`
	Executable  string        `json:"exe,omitempty" journald:"_EXE"`
	CommandLine string        `json:"cmdLine,omitempty" journald:"_CMDLINE"`
	SystemdUnit string        `json:"systemdUnit,omitempty" journald:"_SYSTEMD_UNIT"`
	BootId      string        `json:"bootId,omitempty" journald:"_BOOT_ID"`
	MachineId   string        `json:"machineId,omitempty" journald:"_MACHINE_ID"`
	Hostname    string        `json:"hostname,omitempty" journald:"_HOSTNAME"`
	Transport   string        `json:"transport,omitempty" journald:"_TRANSPORT"`
	Priority    Priority      `json:"priority" journald:"PRIORITY"`
	Message     string        `json:"message" journald:"MESSAGE"`
	MessageId   string        `json:"messageId,omitempty" journald:"MESSAGE_ID"`
	Errno       int           `json:"machineId,omitempty" journald:"ERRNO"`
	SeqId       int64         `json:"seq,omitempty" `
	Facility    int           `json:"syslogFacility,omitempty" journald:"SYSLOG_FACILITY"`
	Identifier  string        `json:"syslogIdent,omitempty" journald:"SYSLOG_IDENTIFIER"`
	SysPID      int           `json:"syslogPid,omitempty" journald:"SYSLOG_PID"`
	Device      string        `json:"kernelDevice,omitempty" journald:"_KERNEL_DEVICE"`
	Subsystem   string        `json:"kernelSubsystem,omitempty" journald:"_KERNEL_SUBSYSTEM"`
	SysName     string        `json:"kernelSysName,omitempty" journald:"_UDEV_SYSNAME"`
	DevNode     string        `json:"kernelDevNode,omitempty" journald:"_UDEV_DEVNODE"`
}

func NewRecord(journal Journal, logger *Logger, config *Config) (*Record, error) {
	record := &Record{}

	err := decodeRecord(journal, reflect.ValueOf(record).Elem(), logger, config)

	if record.TimeUsec == 0 {

		timestamp, err := journal.GetRealtimeUsec()
		if err != nil {
			logger.Error.Printf("Unable to read the time : %s %v", err.Error(), err)
			record.TimeUsec = time.Now().Unix() * 1000
		} else {
			record.TimeUsec = int64(timestamp / 1000)
		}
	}

	return record, err
}

func decodeRecord(journal Journal, toVal reflect.Value, logger *Logger, config *Config) error {
	toType := toVal.Type()

	numField := toVal.NumField()

	for i := 0; i < numField; i++ {
		fieldVal := toVal.Field(i)
		fieldDef := toType.Field(i)
		fieldType := fieldDef.Type
		fieldTag := fieldDef.Tag
		fieldTypeKind := fieldType.Kind()

		jdKey := fieldTag.Get("journald")
		if jdKey == "" {
			continue
		}

		if (!config.AllowField(jdKey)) {
			continue
		}

		value, err := journal.GetDataValue(jdKey)
		if err != nil || value == "" {
			fieldVal.Set(reflect.Zero(fieldType))
			continue
		}

		switch fieldTypeKind {
		case reflect.Int:
			intVal, err := strconv.Atoi(value)
			if err != nil {
				logger.Warning.Printf("Can't convert field %s to int", jdKey)
				fieldVal.Set(reflect.Zero(fieldType))
				continue
			}
			fieldVal.SetInt(int64(intVal))
			break
		case reflect.String:

			fieldVal.SetString(trimField(value, config.FieldLength))
			break

		case reflect.Int64:
			u, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				logger.Warning.Printf("Can't convert field %s to int64", jdKey)
				fieldVal.Set(reflect.Zero(fieldType))
				continue
			}
			fieldVal.SetInt(u / 1000)
			break

		default:
			logger.Warning.Printf("Can't convert field %s unsupported type %s", jdKey, fieldTypeKind)
		}
	}

	return nil
}
func trimField(value string, fieldLength int) string {

	if (fieldLength == 0) {
		fieldLength = 255
	}

	if fieldLength < len(value) {
		return value[0:fieldLength]
	} else {
		return value
	}
}