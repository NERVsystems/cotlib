package cotlib

import (
	"context"
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

func TestLoggerFromContextWithNilLogger(t *testing.T) {
	prev := logger.Load()
	defer SetLogger(prev)

	SetLogger(nil)

	l := LoggerFromContext(context.Background())
	if l == nil {
		t.Fatal("LoggerFromContext returned nil")
	}

	l.Info("test message")
}
