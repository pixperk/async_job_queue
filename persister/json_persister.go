package persister

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pixperk/async_job_queue/job"
)

type JSONPersister struct {
	dir string
	mu  sync.Mutex
}

func NewJSONPersister(dir string) *JSONPersister {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(fmt.Sprintf("Failed to create job directory: %v", err))
	}

	return &JSONPersister{dir: dir}
}

func (p *JSONPersister) SaveJob(job *job.TrackedJob) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	filePath := filepath.Join(p.dir, fmt.Sprintf("%s.json", job.JobID))

	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write job file: %w", err)
	}

	return nil
}

func (p *JSONPersister) UpdateStatus(jobID string, status job.JobStatus, retry int, lastErr error) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	filePath := filepath.Join(p.dir, fmt.Sprintf("%s.json", jobID))

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("couldn't read job file: %w", err)
	}

	var j job.TrackedJob
	if err := json.Unmarshal(data, &j); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	j.Status = status
	j.RetryCount = retry
	j.LastError = lastErr

	return p.SaveJob(&j)
}

func (p *JSONPersister) LoadPendingJobs() ([]*job.TrackedJob, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	var jobs []*job.TrackedJob

	files, err := os.ReadDir(p.dir)
	if err != nil {
		return nil, fmt.Errorf("read dir error: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(p.dir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue // skip corrupted files
		}

		var j job.TrackedJob
		if err := json.Unmarshal(data, &j); err != nil {
			continue
		}

		if j.Status == job.StatusPending || j.Status == job.StatusFailed {
			jobs = append(jobs, &j)
		}

	}

	return jobs, nil
}
