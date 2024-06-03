package common

type LogEntry struct {
	Term    int    `json:"term"`
	Command string `json:"command"`
}
