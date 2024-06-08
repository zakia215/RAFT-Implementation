package common

type LogEntry struct {
	Term    int         `json:"term"`
	Command ExecuteArgs `json:"command"`
}

type ExecuteArgs struct {
	Command string `json:"command"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

type ExecuteReply struct {
	Response   string
	LeaderAdd  string
	LeaderPort string
}
