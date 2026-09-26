package job

import jobdomain "MoneyHook/MoneyHook-API/job"

type Handler struct {
	jobsStore jobdomain.Store
	jobName   string
}

func New(jobStore jobdomain.Store, jobName string) *Handler {
	return &Handler{jobsStore: jobStore, jobName: jobName}
}
