package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ryanpramasyana/gdc/internal/package/task/ports"
	"github.com/ryanpramasyana/gdc/internal/shared/apperr"
	"github.com/ryanpramasyana/gdc/test/domain"
)

// TestTaskCreateSequentialIdempotency proves that replaying a request with the
// same Idempotency-Key returns the original response and creates exactly one
// task. The cases are declared as a map (name -> case) rather than a slice.
func TestTaskCreateSequentialIdempotency(t *testing.T) {
	key := uuid.NewString()
	validInput := ports.CreateTaskInput{Title: "Write report"}

	cases := map[string]struct {
		requests    int
		key         string
		input       ports.CreateTaskInput
		wantStatus  int
		wantErrCode string
		wantTasks   int
	}{
		"first request creates a task": {
			requests: 1, key: key, input: validInput, wantStatus: 201, wantTasks: 1,
		},
		"replayed requests return the cached response": {
			requests: 4, key: key, input: validInput, wantStatus: 201, wantTasks: 1,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newTaskFixture(t)

			var firstBody []byte
			for i := 0; i < tc.requests; i++ {
				result, err := fixture.uc.Create(context.Background(), fixture.owner.ID, tc.key, tc.input, renderTask)
				require.NoError(t, err)
				require.Equal(t, tc.wantStatus, result.Status)
				if i == 0 {
					firstBody = result.Body
					require.False(t, result.Replayed)
					continue
				}
				require.True(t, result.Replayed)
				require.Equal(t, firstBody, result.Body)
			}
			require.Equal(t, tc.wantTasks, fixture.store.TaskCount())
		})
	}
}

// TestTaskCreateValidation covers the request guards using a map of cases.
func TestTaskCreateValidation(t *testing.T) {
	cases := map[string]struct {
		key         string
		input       ports.CreateTaskInput
		wantErrCode string
	}{
		"missing idempotency key": {
			key: "", input: ports.CreateTaskInput{Title: "ok"}, wantErrCode: apperr.CodeIdempotencyRequired,
		},
		"malformed idempotency key": {
			key: "not-a-uuid", input: ports.CreateTaskInput{Title: "ok"}, wantErrCode: apperr.CodeIdempotencyInvalid,
		},
		"empty title": {
			key: uuid.NewString(), input: ports.CreateTaskInput{Title: "   "}, wantErrCode: apperr.CodeValidation,
		},
		"invalid status": {
			key: uuid.NewString(), input: ports.CreateTaskInput{Title: "ok", Status: "nope"}, wantErrCode: apperr.CodeValidation,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newTaskFixture(t)
			_, err := fixture.uc.Create(context.Background(), fixture.owner.ID, tc.key, tc.input, renderTask)

			require.Error(t, err)
			require.Equal(t, tc.wantErrCode, apperr.From(err).Code)
			require.Equal(t, 0, fixture.store.TaskCount())
		})
	}
}

// TestTaskListFilterSearchPagination exercises listing behavior with a map of
// independent query cases.
func TestTaskListFilterSearchPagination(t *testing.T) {
	fixture := newTaskFixture(t)
	ctx := context.Background()

	seed := []ports.CreateTaskInput{
		{Title: "Alpha report", Status: domain.TaskStatusTodo},
		{Title: "Beta report", Status: domain.TaskStatusDone},
		{Title: "Gamma task", Status: domain.TaskStatusTodo},
	}
	for _, input := range seed {
		_, err := fixture.uc.Create(ctx, fixture.owner.ID, uuid.NewString(), input, renderTask)
		require.NoError(t, err)
	}

	cases := map[string]struct {
		input     ports.ListTasksInput
		wantTotal int64
		wantItems int
	}{
		"all tasks": {
			input: ports.ListTasksInput{}, wantTotal: 3, wantItems: 3,
		},
		"filter by status": {
			input: ports.ListTasksInput{Status: domain.TaskStatusTodo}, wantTotal: 2, wantItems: 2,
		},
		"search by title": {
			input: ports.ListTasksInput{Search: "report"}, wantTotal: 2, wantItems: 2,
		},
		"search and filter combined": {
			input: ports.ListTasksInput{Search: "report", Status: domain.TaskStatusDone}, wantTotal: 1, wantItems: 1,
		},
		"pagination limits items": {
			input: ports.ListTasksInput{Page: 1, Limit: 2}, wantTotal: 3, wantItems: 2,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tasks, total, err := fixture.uc.List(ctx, fixture.owner.ID, tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.wantTotal, total)
			require.Len(t, tasks, tc.wantItems)
		})
	}
}

// TestTaskAccessControl verifies that a non-owner cannot read or mutate a task.
func TestTaskAccessControl(t *testing.T) {
	fixture := newTaskFixture(t)
	ctx := context.Background()
	stranger := fixture.addTeammate(t, "stranger@example.com")

	result, err := fixture.uc.Create(ctx, fixture.owner.ID, uuid.NewString(), ports.CreateTaskInput{Title: "Private"}, renderTask)
	require.NoError(t, err)
	require.Equal(t, 201, result.Status)
	taskID := uint64(1)

	cases := map[string]struct {
		action      func() error
		wantErrCode string
	}{
		"stranger cannot get": {
			action:      func() error { _, err := fixture.uc.Get(ctx, stranger.ID, taskID); return err },
			wantErrCode: apperr.CodeForbidden,
		},
		"stranger cannot update": {
			action: func() error {
				_, err := fixture.uc.Update(ctx, stranger.ID, taskID, ports.UpdateTaskInput{Title: "x"})
				return err
			},
			wantErrCode: apperr.CodeForbidden,
		},
		"stranger cannot delete": {
			action:      func() error { return fixture.uc.Delete(ctx, stranger.ID, taskID) },
			wantErrCode: apperr.CodeForbidden,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.wantErrCode, apperr.From(tc.action()).Code)
		})
	}
}
