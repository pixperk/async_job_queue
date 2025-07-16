package job

import (
	"context"
	"fmt"
	"time"
)

type SleepyJob struct {
	Duration time.Duration
}

func (j *SleepyJob) Execute(ctx context.Context) error {
	select {
	case <-time.After(j.Duration):
		fmt.Println("😴 Job finished sleeping")
		return nil
	case <-ctx.Done():
		return ctx.Err() // return context.Canceled or context.DeadlineExceeded
	}
}
