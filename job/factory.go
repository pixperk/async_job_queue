package job

import (
	"encoding/json"
	"fmt"

	trackedjob "github.com/pixperk/async_job_queue/trackedJob"
)

type JobCreator func(config json.RawMessage) (trackedjob.Job, error)

type JobFactory struct {
	creators map[string]JobCreator
}

func NewJobFactory() *JobFactory {
	return &JobFactory{
		creators: make(map[string]JobCreator),
	}
}

func (f *JobFactory) Register(jobType string, creator JobCreator) {
	f.creators[jobType] = creator
}

func (f *JobFactory) Create(jobType string, config json.RawMessage) (trackedjob.Job, error) {
	creator, ok := f.creators[jobType]
	if !ok {
		return nil, fmt.Errorf("no creator registered for job type: %s", jobType)
	}
	return creator(config)
}
