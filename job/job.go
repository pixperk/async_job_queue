package job

import (
	"fmt"
	"sync"
)

type Job interface {
	Execute() error
}

type JobQueue struct {
	jobs    chan *TrackedJob
	workers int
	wg      sync.WaitGroup
	tracker *JobTracker
}

func NewJobQueue(buffer int, workers int) *JobQueue {
	q := &JobQueue{
		jobs:    make(chan *TrackedJob, buffer),
		workers: workers,
		tracker: NewJobTracker(),
	}

	for i := 0; i < q.workers; i++ {
		go func(workerID int) {
			for job := range q.jobs {
				fmt.Printf("👷 Worker %d picked a job\n", workerID)
				job.ExecuteWithRetry()
				q.wg.Done()
			}
		}(i + 1)
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
