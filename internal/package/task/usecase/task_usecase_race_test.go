package usecase_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ryanpramasyana/gdc/internal/package/task/ports"
)

// TestTaskCreateConcurrentIdempotency is the core race-condition test: many
// goroutines submit the same Idempotency-Key simultaneously and exactly one
// task must be created, with every caller receiving the identical response.
// Run with `go test -race`.
func TestTaskCreateConcurrentIdempotency(t *testing.T) {
	const goroutines = 64

	fixture := newTaskFixture(t)
	key := uuid.NewString()
	input := ports.CreateTaskInput{Title: "Concurrent create"}

	var (
		wg     sync.WaitGroup
		start  = make(chan struct{})
		bodies = make([][]byte, goroutines)
		errs   = make([]error, goroutines)
	)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			result, err := fixture.uc.Create(context.Background(), fixture.owner.ID, key, input, renderTask)
			if err != nil {
				errs[index] = err
				return
			}
			bodies[index] = result.Body
		}(i)
	}

	close(start)
	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "goroutine %d returned an error", i)
	}
	require.Equal(t, 1, fixture.store.TaskCount(), "exactly one task must be created")
	for i := 1; i < goroutines; i++ {
		require.Equal(t, bodies[0], bodies[i], "goroutine %d received a different response", i)
	}
}
