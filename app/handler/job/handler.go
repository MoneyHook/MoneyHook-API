package job

import jobdomain "MoneyHook/MoneyHook-API/job"

type Handler struct {
	jobsStore jobdomain.Store
}

func New(jobStore jobdomain.Store) *Handler {
	return &Handler{jobsStore: jobStore}
}
