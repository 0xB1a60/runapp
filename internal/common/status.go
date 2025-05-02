package common

type AppStatus string

const (
	AppStatusStarting AppStatus = "starting"
	AppStatusRunning  AppStatus = "running"
	AppStatusSuccess  AppStatus = "success"
	AppStatusFailed   AppStatus = "failed"
)

var AppStatusPretty = map[AppStatus]string{
	AppStatusStarting: "Starting",
	AppStatusRunning:  "Running",
	AppStatusSuccess:  "Success",
	AppStatusFailed:   "Failed",
}

var AppStatusPriority = map[AppStatus]int{
	AppStatusStarting: 1,
	AppStatusRunning:  2,
	AppStatusSuccess:  3,
	AppStatusFailed:   4,
}
