package job

import (
	"crypto/rand"
	"fmt"
)

func GenerateJobID() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return "job-xxxx"
	}
	return fmt.Sprintf("job-%x", b)
}
