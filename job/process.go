package job

import (
	"context"
	"fmt"
)

func ProcessJob(ctx context.Context, q *JobQueue, workerID int) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d shutting down: %v\n", workerID, ctx.Err())
			return
		case job, ok := <-q.jobs:
			if !ok {
				fmt.Printf(" Job channel closed. Worker %d exiting\n", workerID)
				return
			}
			fmt.Printf("👷 Worker %d picked a job\n", workerID)
			job.ExecuteWithRetry(ctx)
			q.wg.Done()
		}
	}
}
