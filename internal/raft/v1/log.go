package v1

type LogEntry struct {

	// 当前的日志条目所属任期
	CurrTerm uint64

	// 当前日志条目索引
	Index int

	// 客户端命令
	LogCmd interface{}
}
