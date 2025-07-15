package job

import (
	"fmt"
	"time"
)

type SampleJob struct {
	ID int
}

func (s *SampleJob) Execute() error {
	startTime := time.Now()
	fmt.Printf("🔧 Job %d started at %s\n", s.ID, startTime.Format("15:04:05.000"))

	time.Sleep(2 * time.Second) // simulate slowness

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Printf("✅ Job %d finished at %s (duration: %v)\n", s.ID, endTime.Format("15:04:05.000"), duration)
	return nil
}

type ErroneousJob struct {
	ID       int
	Attempts int
}

func (f *ErroneousJob) Execute() error {
	f.Attempts++
	fmt.Printf("Executing ErroneousJob %d (attempt %d)\n", f.ID, f.Attempts)

	if f.Attempts < 3 {
		return fmt.Errorf("simulated failure on attempt %d", f.Attempts)
	}

	time.Sleep(1 * time.Second)
	fmt.Printf("🎉 ErroneousJob %d succeeded\n", f.ID)
	return nil
}
