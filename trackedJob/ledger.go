package trackedjob

import (
	"fmt"
	"sync"
	"time"
)

type JobTracker struct {
	mu    sync.RWMutex
	store map[string]*JobMetadata
}

type JobMetadata struct {
	JobID      string
	Status     JobStatus
	RetryCount int
	LastError  error
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewJobTracker() *JobTracker {
	return &JobTracker{
		store: make(map[string]*JobMetadata),
	}
}

func (jt *JobTracker) Register(jobID string) {
	jt.mu.Lock()
	defer jt.mu.Unlock()
	jt.store[jobID] = &JobMetadata{
		JobID:     jobID,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (jt *JobTracker) Update(jobID string, status JobStatus, retry int, err error) {
	jt.mu.Lock()
	defer jt.mu.Unlock()
	if meta, exists := jt.store[jobID]; exists {
		meta.Status = status
		meta.RetryCount = retry
		meta.LastError = err
		meta.UpdatedAt = time.Now()
	}
}

func (jt *JobTracker) PrintAll() {
	jt.mu.RLock()
	defer jt.mu.RUnlock()
	fmt.Println("\n Job Tracker Status:")
	for _, meta := range jt.store {
		fmt.Printf("- [%s] JobID: %s | Retries: %d | Error: %v | Updated: %s\n",
			meta.Status, meta.JobID, meta.RetryCount, meta.LastError, meta.UpdatedAt.Format("15:04:05"))
	}
}
