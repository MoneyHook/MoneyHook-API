package model

type Status string

const (
	UserAgent                   = "User-Agent"
	XCloudScheduler             = "X-CloudScheduler"
	XCloudSchedulerJobName      = "X-CloudScheduler-JobName"
	XCloudSchedulerScheduleTime = "X-CloudScheduler-ScheduleTime"
	ContentType                 = "Content-Type"
)

const (
	Success = Status("Success")
	Error   = Status("Error")
)

func (s Status) Create(message *string) map[string]string {

	switch s {
	case Success:
		return map[string]string{"status": "success"}
	case Error:
		return map[string]string{"status": "error", "message": *message}
	default:
		return map[string]string{"status": "success"}
	}
}
