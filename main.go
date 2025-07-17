package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pixperk/async_job_queue/job"
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

	q := job.NewJobQueue(ctx, buffer, workers)

	q.Submit(&job.SleepyJob{Duration: 1 * time.Second}, job.SubmitOptions{
		MaxRetries: 2,
		Timeout:    4 * time.Second,
	})
	q.Submit(&job.SleepyJob{Duration: 6 * time.Second}, job.SubmitOptions{
		MaxRetries: 0,
	})

	q.Wait()
	q.Shutdown()

	q.StoreStatus()
	fmt.Printf("\nTotal execution time: %.2fs\n", time.Since(start).Seconds())
}
