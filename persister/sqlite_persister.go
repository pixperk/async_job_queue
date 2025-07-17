package persister

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	trackedjob "github.com/pixperk/async_job_queue/trackedJob"
)

type SQLitePersister struct {
	db *sql.DB
	mu sync.Mutex
}

func NewSQLitePersister(path string) (*SQLitePersister, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// Ensure table exists
	schema := `
	CREATE TABLE IF NOT EXISTS jobs (
		job_id TEXT PRIMARY KEY,
		job_type TEXT NOT NULL,
		job_data TEXT NOT NULL,
		status TEXT NOT NULL,
		retry_count INTEGER,
		max_retries INTEGER,
		timeout_seconds INTEGER,
		last_error TEXT
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &SQLitePersister{db: db}, nil
}

func (p *SQLitePersister) SaveJob(job *trackedjob.TrackedJob) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	payload, err := job.Job.Serialize()
	if err != nil {
		return fmt.Errorf("serialization failed: %w", err)
	}

	_, err = p.db.Exec(`
		INSERT OR REPLACE INTO jobs
		(job_id, job_type, job_data, status, retry_count, max_retries, timeout_seconds, last_error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		job.JobID,
		job.JobType,
		payload,
		job.Status,
		job.RetryCount,
		job.MaxRetries,
		int(job.Timeout.Seconds()),
		job.LastError,
	)

	if err != nil {
		return fmt.Errorf("failed to save job %s: %w", job.JobID, err)
	}

	return nil
}

func (p *SQLitePersister) UpdateStatus(jobID string, status trackedjob.JobStatus, retry int, err error) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	_, execErr := p.db.Exec(`
		UPDATE jobs SET status = ?, retry_count = ?, last_error = ?
		WHERE job_id = ?
	`, status, retry, errStr, jobID)

	if execErr != nil {
		return fmt.Errorf("failed to update job %s: %w", jobID, execErr)
	}

	return nil
}

func (p *SQLitePersister) LoadPendingJobs() ([]*trackedjob.TrackedJob, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	rows, err := p.db.Query(`
		SELECT job_id, job_type, job_data, status, retry_count, max_retries, timeout_seconds, last_error
		FROM jobs
		WHERE status = ? OR status = ?
	`, trackedjob.StatusPending, trackedjob.StatusFailed)

	if err != nil {
		return nil, fmt.Errorf("failed to query pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*trackedjob.TrackedJob

	for rows.Next() {
		var (
			jobID      string
			jobType    string
			jobData    []byte
			status     string
			retryCount int
			maxRetries int
			timeoutSec int
			lastError  sql.NullString
		)

		if err := rows.Scan(&jobID, &jobType, &jobData, &status, &retryCount, &maxRetries, &timeoutSec, &lastError); err != nil {
			continue
		}

		tJob := &trackedjob.TrackedJob{
			JobID:      jobID,
			JobType:    jobType,
			JobData:    jobData,
			Status:     trackedjob.JobStatus(status),
			RetryCount: retryCount,
			MaxRetries: maxRetries,
			Timeout:    0,
			LastError:  "",
		}
		if timeoutSec > 0 {
			tJob.Timeout = time.Duration(int64(timeoutSec)) * 1e9 // seconds to nanoseconds
		}
		if lastError.Valid {
			tJob.LastError = lastError.String
		}
		jobs = append(jobs, tJob)
	}

	return jobs, nil
}
