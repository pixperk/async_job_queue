package job

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/pixperk/async_job_queue/persister"
	"github.com/pixperk/async_job_queue/retry"
	trackedjob "github.com/pixperk/async_job_queue/trackedJob"
)

type JobQueue struct {
	jobs      chan *trackedjob.TrackedJob
	workers   int
	wg        sync.WaitGroup
	persister persister.JobPersister
}

func NewJobQueue(ctx context.Context, buffer int, workers int, persister persister.JobPersister) *JobQueue {
	q := &JobQueue{
		jobs:      make(chan *trackedjob.TrackedJob, buffer),
		workers:   workers,
		persister: persister,
	}

	for i := 0; i < q.workers; i++ {
		go ProcessJob(ctx, q, i+1)
	}

	return q
}

func (q *JobQueue) handleStatusUpdate(jobID string, status trackedjob.JobStatus, retry int, err error) {

	if q.persister != nil {
		saveErr := q.persister.UpdateStatus(jobID, status, retry, err)
		if saveErr != nil {
			log.Printf("Persistence update failed for job %s: %v", jobID, saveErr)
		}
	}
}

func (q *JobQueue) Submit(job trackedjob.Job, opts SubmitOptions) {
	q.wg.Add(1)

	jobID := GenerateJobID()
	jobData, err := job.Serialize()
	if err != nil {
		log.Printf("Failed to serialize job %s: %v", jobID, err)
		q.wg.Done()
		return
	}

	tracked := &trackedjob.TrackedJob{
		JobID:          jobID,
		JobType:        job.TypeName(),
		JobData:        jobData,
		Job:            job,
		MaxRetries:     opts.MaxRetries,
		Timeout:        opts.Timeout,
		Status:         trackedjob.StatusPending,
		Metadata:       opts.Metadata,
		OnStatusUpdate: q.handleStatusUpdate,
		Backoff: retry.ExponentialBackoff{
			BaseDelay: 500 * time.Millisecond,
			MaxDelay:  5 * time.Second,
			Jitter:    true,
		},
	}
	//persist the job

	if q.persister != nil {
		if err := q.persister.SaveJob(tracked); err != nil {
			log.Printf("Failed to persist job %s: %v", jobID, err)
		}
	}

	q.jobs <- tracked
}

func (q *JobQueue) LoadPersistedJobs(factory *JobFactory) {
	if q.persister == nil {
		return
	}
	jobs, err := q.persister.LoadPendingJobs()
	if err != nil {
		log.Printf("Failed to load persisted jobs: %v", err)
		return
	}
	for _, j := range jobs {
		rehydratedJob, err := factory.Create(j.JobType, j.JobData)
		if err != nil {
			log.Printf("Failed to rehydrate job %s of type %s: %v", j.JobID, j.JobType, err)
			continue
		}
		j.Job = rehydratedJob
		j.OnStatusUpdate = q.handleStatusUpdate
		j.Backoff = retry.ExponentialBackoff{
			BaseDelay: 500 * time.Millisecond,
			MaxDelay:  5 * time.Second,
			Jitter:    true,
		}

		log.Printf("Re-queuing persisted jobs: %s", j.JobID)
		q.wg.Add(1)
		q.jobs <- j
	}
}

type SubmitOptions struct {
	MaxRetries int
	Timeout    time.Duration
	Metadata   map[string]string
}

func (q *JobQueue) Wait() {
	q.wg.Wait()
}

func (q *JobQueue) Shutdown() {
	close(q.jobs)
}
