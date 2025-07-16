package job

import (
	"fmt"
	"sync"
	"time"
)

type TrackedJob struct {
	JobID      string
	Job        Job //wrap job interface
	RetryCount int
	MaxRetries int
	Status     JobStatus
	LastError  error
	Tracker    *JobTracker
	mu         sync.Mutex
}

type JobStatus string

const (
	StatusPending JobStatus = "PENDING"
	StatusRunning JobStatus = "RUNNING"
	StatusFailed  JobStatus = "FAILED"
	StatusSuccess JobStatus = "SUCCESS"
)

func (t *TrackedJob) ExecuteWithRetry() {
	t.mu.Lock()
	t.Status = StatusRunning
	//Update ledger
	t.Tracker.Update(t.JobID, StatusRunning, 0, nil)
	t.mu.Unlock()

	for attempt := 1; attempt <= t.MaxRetries+1; attempt++ {
		err := t.Job.Execute()
		if err != nil {
			//track the job
			t.mu.Lock()
			t.LastError = err
			t.Status = StatusFailed
			t.RetryCount = attempt
			//Update ledger
			t.Tracker.Update(t.JobID, StatusFailed, attempt, err)
			t.mu.Unlock()

			fmt.Printf("Attempt %d failed: %v\n", attempt, err)
			if attempt <= t.MaxRetries {
				fmt.Printf("Retrying job (attempt %d)...\n", attempt+1)
				time.Sleep(500 * time.Millisecond) // backoff delay
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
