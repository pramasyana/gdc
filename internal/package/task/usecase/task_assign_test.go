package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ryanpramasyana/gdc/internal/shared/apperr"
	"github.com/ryanpramasyana/gdc/test/domain"
)

// createTask is a helper that creates a task owned by the fixture owner.
func (f *taskFixture) createTask(t *testing.T, title string) uint64 {
	t.Helper()
	result, err := f.uc.Create(context.Background(), f.owner.ID, uuid.NewString(), createInput(title), renderTask)
	require.NoError(t, err)
	var payload struct {
		ID uint64 `json:"id"`
	}
	require.NoError(t, jsonUnmarshal(result.Body, &payload))
	return payload.ID
}

// TestTaskAssignSuccess verifies the happy path: assignee updated, audit log
// written and notification dispatched.
func TestTaskAssignSuccess(t *testing.T) {
	fixture := newTaskFixture(t)
	teammate := fixture.addTeammate(t, "member@example.com")
	taskID := fixture.createTask(t, "Assign me")

	task, err := fixture.uc.Assign(context.Background(), fixture.owner.ID, taskID, teammate.ID)
	require.NoError(t, err)
	require.NotNil(t, task.AssigneeID)
	require.Equal(t, teammate.ID, *task.AssigneeID)

	assignee, ok := fixture.store.TaskAssignee(taskID)
	require.True(t, ok)
	require.Equal(t, teammate.ID, *assignee)
	require.Equal(t, 1, fixture.store.TaskLogCount())
	require.Len(t, fixture.store.Notifications(), 1)
}

// TestTaskAssignRejections covers the assignment guards with a map of cases.
func TestTaskAssignRejections(t *testing.T) {
	cases := map[string]struct {
		wantErrCode string
	}{
		"non-owner cannot assign":   {wantErrCode: apperr.CodeForbidden},
		"assignee outside the team": {wantErrCode: apperr.CodeUnprocessable},
		"owner cannot assign self":  {wantErrCode: apperr.CodeUnprocessable},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newTaskFixture(t)
			member := fixture.addTeammate(t, "member@example.com")
			outsider := fixture.store.AddUser(domain.User{Name: "Outsider", Email: "outsider@example.com"})
			taskID := fixture.createTask(t, "Assign me")

			var err error
			switch name {
			case "non-owner cannot assign":
				err = assignErr(fixture, member.ID, taskID, member.ID)
			case "assignee outside the team":
				err = assignErr(fixture, fixture.owner.ID, taskID, outsider.ID)
			case "owner cannot assign self":
				err = assignErr(fixture, fixture.owner.ID, taskID, fixture.owner.ID)
			}

			require.Equal(t, tc.wantErrCode, apperr.From(err).Code)
			_, assigned := fixture.store.TaskAssignee(taskID)
			require.False(t, assigned, "no assignee should be persisted on rejection")
			require.Equal(t, 0, fixture.store.TaskLogCount())
		})
	}
}

// TestTaskAssignRollback verifies transactional integrity: when notification
// fails the assignee update and audit log are rolled back.
func TestTaskAssignRollback(t *testing.T) {
	fixture := newTaskFixture(t)
	teammate := fixture.addTeammate(t, "member@example.com")
	taskID := fixture.createTask(t, "Assign me")

	fixture.store.SetNotifyError(errors.New("notification transport down"))

	err := assignErr(fixture, fixture.owner.ID, taskID, teammate.ID)
	require.Error(t, err)

	_, assigned := fixture.store.TaskAssignee(taskID)
	require.False(t, assigned, "assignee update must be rolled back")
	require.Equal(t, 0, fixture.store.TaskLogCount(), "audit log must be rolled back")
}

func assignErr(fixture *taskFixture, actorID, taskID, assigneeID uint64) error {
	_, err := fixture.uc.Assign(context.Background(), actorID, taskID, assigneeID)
	return err
}
