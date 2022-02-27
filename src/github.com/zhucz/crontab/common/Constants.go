package common

const (
	JobSaveDir     = "/cron/jobs/"
	JobKillerDir   = "/cron/killer/"
	JobLockDir     = "/cron/lock/"
	JobWorkerDir   = "/cron/workers/"
	JobEventPut    = 1
	JobEventDelete = 2
	JobEventKiller = 3
)
