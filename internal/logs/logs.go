package logs

type LogType = string

const (
	AllLogs LogType = "all"
	OutLogs LogType = "stdout"
	ErrLogs LogType = "stderr"
)

var ValidTypes = []LogType{
	AllLogs,
	OutLogs,
	ErrLogs,
}

const (
	maxBufferCapacity = 10 * 1024 * 1024 // 10MB
)
