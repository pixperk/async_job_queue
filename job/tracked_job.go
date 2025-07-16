package job

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pixperk/async_job_queue/retry"
)

type TrackedJob struct {
	JobID      string
	Job        Job //wrap job interface
	RetryCount int
	MaxRetries int
	Status     JobStatus
	LastError  error
	Tracker    *JobTracker
	Backoff    retry.BackoffStrategy
	Timeout    time.Duration
	mu         sync.Mutex
}

type JobStatus string

const (
	StatusPending JobStatus = "PENDING"
	StatusRunning JobStatus = "RUNNING"
	StatusFailed  JobStatus = "FAILED"
	StatusSuccess JobStatus = "SUCCESS"
)

func (t *TrackedJob) ExecuteWithRetry(ctx context.Context) {
	t.mu.Lock()
	t.Status = StatusRunning
	//Update ledger
	t.Tracker.Update(t.JobID, StatusRunning, 0, nil)
	t.mu.Unlock()

	for attempt := 1; attempt <= t.MaxRetries+1; attempt++ {
		// Derive a child context with timeout for this attempt
		attemptTimeout := t.Timeout
		if attemptTimeout == 0 {
			attemptTimeout = 5 * time.Second // default fallback
		}
		attemptCtx, cancel := context.WithTimeout(ctx, attemptTimeout)
		defer cancel()
		err := t.Job.Execute(attemptCtx)

		if err != nil {
			t.mu.Lock()
			//track the job
			t.LastError = err
			t.Status = StatusFailed
			t.RetryCount = attempt
			//Update ledger
			t.Tracker.Update(t.JobID, StatusFailed, attempt, err)
			t.mu.Unlock()

			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Job attempt %d timed out\n", attempt)
			} else {
				fmt.Printf("Attempt %d failed: %v\n", attempt, err)
			}

			fmt.Printf("Attempt %d failed: %v\n", attempt, err)
			if attempt <= t.MaxRetries {
				delay := t.Backoff.NextDelay(attempt)
				fmt.Printf("Retrying job (attempt %d)...\n", attempt+1)
				time.Sleep(delay) // backoff delay
				continue
			} else {
				fmt.Printf("Job permanently failed after %d attempts.\n", attempt-1)
				return
			}
		} else {
			t.mu.Lock()
			//Update ledger
			t.Tracker.Update(t.JobID, StatusSuccess, attempt, nil)
			t.Status = StatusSuccess
			t.mu.Unlock()

			fmt.Printf("✅ Job succeeded on attempt %d\n", attempt)
			return
		}
	}
}
