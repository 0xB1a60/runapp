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
