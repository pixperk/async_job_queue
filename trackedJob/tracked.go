package trackedjob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pixperk/async_job_queue/retry"
)

type Job interface {
	Execute(ctx context.Context) error
	TypeName() string
	Serialize() (json.RawMessage, error)
}

type TrackedJob struct {
	JobID      string                `json:"job_id"`
	JobType    string                `json:"job_type"`
	JobData    json.RawMessage       `json:"job_data"`
	Job        Job                   `json:"-"`
	RetryCount int                   `json:"retry_count"`
	MaxRetries int                   `json:"max_retries"`
	Status     JobStatus             `json:"status"`
	Metadata   map[string]string     `json:"metadata"`
	LastError  string                `json:"last_error"`
	Backoff    retry.BackoffStrategy `json:"-"`
	Timeout    time.Duration
	mu         sync.Mutex `json:"-"`

	OnStatusUpdate func(jobID string, status JobStatus, retry int, err error) `json:"-"`
}

type JobStatus string

const (
	StatusPending JobStatus = "PENDING"
	StatusRunning JobStatus = "RUNNING"
	StatusFailed  JobStatus = "FAILED"
	StatusSuccess JobStatus = "SUCCESS"
)

func (t *TrackedJob) ExecuteWithRetry(ctx context.Context) {
	t.updateStatus(StatusRunning, 0, nil)

	for attempt := 1; attempt <= t.MaxRetries+1; attempt++ {
		// Derive a child context with timeout for this attempt
		attemptTimeout := t.Timeout
		if attemptTimeout == 0 {
			attemptTimeout = 5 * time.Second // default fallback
		}
		attemptCtx, cancel := context.WithTimeout(ctx, attemptTimeout)

		err := t.Job.Execute(attemptCtx)

		cancel()

		if err != nil {
			t.updateStatus(StatusFailed, attempt, err)

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
			t.updateStatus(StatusSuccess, attempt, nil)
			fmt.Printf("✅ Job succeeded on attempt %d\n", attempt)
			return
		}
	}
}

func (t *TrackedJob) updateStatus(status JobStatus, attempt int, err error) {
	t.mu.Lock()
	t.Status = status
	t.RetryCount = attempt
	if err != nil {
		t.LastError = err.Error()
	} else {
		t.LastError = ""
	}
	t.mu.Unlock()

	if t.OnStatusUpdate != nil {
		t.OnStatusUpdate(t.JobID, status, attempt, err)
	}
}

func (t *TrackedJob) StatusString() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return string(t.Status)
}
