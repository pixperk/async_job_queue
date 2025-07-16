package main

import (
	"fmt"
	"time"

	"github.com/pixperk/async_job_queue/job"
)

func main() {
	start := time.Now()

	buffer, workers := 10, 3

	q := job.NewJobQueue(buffer, workers)

	jobNumber := 5

	for range jobNumber {
		maxRetries := jobNumber
		q.Submit(&job.ErroneousJob{ID: 2}, maxRetries)
	}

	q.Wait()
	q.Shutdown()

	q.StoreStatus()
	fmt.Printf("\nTotal execution time: %.2fs\n", time.Since(start).Seconds())
}
