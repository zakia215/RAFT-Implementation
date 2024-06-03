package common

type LogEntry struct {
	Term    int         `json:"term"`
	Command interface{} `json:"command"`
}

type ExecuteArgs struct {
	Command string
	Key     string
	Value   string
}

type ExecuteReply struct {
	Response string
}
