package persister

import "github.com/pixperk/async_job_queue/job"

type jobPersister interface {
	SaveJob(job *job.TrackedJob) error
	UpdateStatus(jobID string, status job.JobStatus, retry int, err error) error
	LoadPendingJobs() ([]*job.TrackedJob, error)
}
