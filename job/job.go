package job

import (
	"context"
	"sync"
	"time"

	"github.com/pixperk/async_job_queue/retry"
)

type Job interface {
	Execute(ctx context.Context) error
}

type JobQueue struct {
	jobs    chan *TrackedJob
	workers int
	wg      sync.WaitGroup
	tracker *JobTracker
}

func NewJobQueue(ctx context.Context, buffer int, workers int) *JobQueue {
	q := &JobQueue{
		jobs:    make(chan *TrackedJob, buffer),
		workers: workers,
		tracker: NewJobTracker(),
	}

	for i := 0; i < q.workers; i++ {
		go ProcessJob(ctx, q, i+1)
	}

	return q
}

func (q *JobQueue) Submit(job Job, maxRetries int) {
	q.wg.Add(1)

	jobID := GenerateJobID()

	tracked := &TrackedJob{
		JobID:      jobID,
		Job:        job,
		MaxRetries: maxRetries,
		Status:     StatusPending,
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

func (q *JobQueue) Wait() {
	q.wg.Wait()
}

func (q *JobQueue) Shutdown() {
	close(q.jobs)
}

func (q *JobQueue) StoreStatus() {
	q.tracker.PrintAll()
}

func (t *TrackedJob) StatusString() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return string(t.Status)
}
