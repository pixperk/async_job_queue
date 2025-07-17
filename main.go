package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pixperk/async_job_queue/job"
	"github.com/pixperk/async_job_queue/persister"
	trackedjob "github.com/pixperk/async_job_queue/trackedJob"
)

func main() {
	start := time.Now()

	buffer, workers := 10, 3

	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %s. Shutting down gracefully...\n", sig)
		cancel() // Broadcast shutdown to all workers
	}()

	persister, err := persister.NewSQLitePersister("./saved_jobs/jobs.db")
	if err != nil {
		panic(err)
	}
	jobFactory := job.NewJobFactory()
	jobFactory.Register("SleepyJob", func(config json.RawMessage) (trackedjob.Job, error) {
		var sj job.SleepyJob
		if err := json.Unmarshal(config, &sj); err != nil {
			return nil, err
		}
		return &sj, nil
	})

	q := job.NewJobQueue(ctx, buffer, workers, persister)

	q.LoadPersistedJobs(jobFactory)

	/* q.Submit(&job.SleepyJob{Duration: 1 * time.Second}, job.SubmitOptions{
		MaxRetries: 2,
		Timeout:    4 * time.Second,
	})
	q.Submit(&job.SleepyJob{Duration: 6 * time.Second}, job.SubmitOptions{
		MaxRetries: 0,
		Timeout:    10 * time.Second, // Give enough time for the 6-second job
	}) */

	q.Wait()
	q.Shutdown()

	fmt.Printf("\nTotal execution time: %.2fs\n", time.Since(start).Seconds())
}
