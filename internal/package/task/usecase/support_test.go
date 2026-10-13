package usecase_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ryanpramasyana/gdc/internal/package/task/ports"
	"github.com/ryanpramasyana/gdc/internal/package/task/usecase"
	"github.com/ryanpramasyana/gdc/test/domain"
	"github.com/ryanpramasyana/gdc/test/mocks"
)

const testIdemTTL = 24 * time.Hour

type taskFixture struct {
	uc    ports.TaskUsecase
	store *mocks.Store
	owner domain.User
	team  domain.Team
}

func newTaskFixture(t *testing.T) *taskFixture {
	t.Helper()
	store := mocks.NewStore()

	owner := store.AddUser(domain.User{Name: "Owner", Email: "owner@example.com", PasswordHash: "x"})
	team := store.AddTeam("Owner team", owner.ID, domain.RoleOwner)
	store.SetPrimaryTeam(owner.ID, team.ID)

	uc := usecase.New(
		mocks.NewTaskRepository(store),
		mocks.NewTaskLogRepository(store),
		mocks.NewUserRepository(store),
		mocks.NewTeamRepository(store),
		mocks.NewIdempotencyRepository(store),
		mocks.NewNotifier(store),
		mocks.NewTxManager(store),
		testIdemTTL,
	)
	return &taskFixture{uc: uc, store: store, owner: owner, team: team}
}

// addTeammate inserts a second user sharing the fixture's team.
func (f *taskFixture) addTeammate(t *testing.T, email string) domain.User {
	t.Helper()
	member := f.store.AddUser(domain.User{Name: "Member", Email: email, PasswordHash: "x"})
	if err := mocks.NewTeamRepository(f.store).AddMember(context.Background(), f.team.ID, member.ID, domain.RoleMember); err != nil {
		t.Fatalf("add teammate: %v", err)
	}
	f.store.SetPrimaryTeam(member.ID, f.team.ID)
	return member
}

// renderTask serializes a task deterministically for idempotency assertions.
func renderTask(task *domain.Task) (status int, body []byte) {
	body, _ = json.Marshal(map[string]any{"id": task.ID, "title": task.Title})
	return 201, body
}

// createInput builds a create-task input with the given title.
func createInput(title string) ports.CreateTaskInput {
	return ports.CreateTaskInput{Title: title}
}

// jsonUnmarshal is a thin indirection so tests avoid importing encoding/json
// directly in every file.
func jsonUnmarshal(data []byte, target any) error {
	return json.Unmarshal(data, target)
}
