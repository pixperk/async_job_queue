package persister

import (
	trackedjob "github.com/pixperk/async_job_queue/trackedJob"
)

type JobPersister interface {
	SaveJob(job *trackedjob.TrackedJob) error
	UpdateStatus(jobID string, status trackedjob.JobStatus, retry int, err error) error
	LoadPendingJobs() ([]*trackedjob.TrackedJob, error)
}
