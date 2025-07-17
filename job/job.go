package job

import (
	"context"
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
	tracker   *trackedjob.JobTracker
	persister persister.JobPersister
}

func NewJobQueue(ctx context.Context, buffer int, workers int) *JobQueue {
	q := &JobQueue{
		jobs:    make(chan *trackedjob.TrackedJob, buffer),
		workers: workers,
		tracker: trackedjob.NewJobTracker(),
	}

	for i := 0; i < q.workers; i++ {
		go ProcessJob(ctx, q, i+1)
	}

	return q
}

func (q *JobQueue) Submit(job trackedjob.Job, opts SubmitOptions) {
	q.wg.Add(1)

	jobID := GenerateJobID()

	tracked := &trackedjob.TrackedJob{
		JobID:      jobID,
		Job:        job,
		MaxRetries: opts.MaxRetries,
		Timeout:    opts.Timeout,
		Status:     trackedjob.StatusPending,
		Metadata:   opts.Metadata,
		Tracker:    q.tracker,
		Backoff: retry.ExponentialBackoff{
			BaseDelay: 500 * time.Millisecond,
			MaxDelay:  5 * time.Second,
			Jitter:    true,
		},
	}

	q.tracker.Register(jobID)

	q.jobs <- tracked
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

func (q *JobQueue) StoreStatus() {
	q.tracker.PrintAll()
}
