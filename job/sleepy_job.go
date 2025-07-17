package job

import (
	"context"
	"encoding/json"
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

func (j *SleepyJob) TypeName() string {
	return "SleepyJob"
}

func (j *SleepyJob) Serialize() (json.RawMessage, error) {
	return json.Marshal(j)
}
