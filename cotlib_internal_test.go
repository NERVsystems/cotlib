package cotlib

import (
	"sync"
	"testing"
)

func TestMaxValueLenRace(t *testing.T) {
	// Run with -race flag to detect races
	var wg sync.WaitGroup
	iterations := 1000
	goroutines := 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := int64(0); j < int64(iterations); j++ {
				// Read maxValueLen
				_ = maxValueLen.Load()

				// Write maxValueLen
				SetMaxValueLen(1024 + j)
			}
		}()
	}

	wg.Wait()
}
