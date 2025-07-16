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

	q.Submit(&job.SleepyJob{Duration: 1 * time.Second}, 0)
	q.Submit(&job.SleepyJob{Duration: 6 * time.Second}, 0)

	q.Wait()
	q.Shutdown()

	q.StoreStatus()
	fmt.Printf("\nTotal execution time: %.2fs\n", time.Since(start).Seconds())
}
