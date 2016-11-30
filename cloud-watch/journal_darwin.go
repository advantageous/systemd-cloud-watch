package cloud_watch

var mockMap = map[string]string{
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

func NewJournal(config *Config) (Journal, error) {
	return NewJournalWithMap(mockMap), nil
}
