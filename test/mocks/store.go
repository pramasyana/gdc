// Package mocks provides concurrency-safe, in-memory implementations of the
// domain ports so usecases can be unit tested without a database or Redis.
package mocks

import (
	"sync"
	"time"

	"github.com/ryanpramasyana/gdc/test/domain"
)

// Store is the shared in-memory state backing every fake repository. All access
// is guarded by a mutex, and it supports snapshot/restore so the transactional
// usecase can be tested including rollback.
type Store struct {
	mu sync.Mutex

	nextUserID uint64
	nextTeamID uint64
	nextTaskID uint64

	users       map[uint64]*domain.User
	emailIndex  map[string]uint64
	teams       map[uint64]*domain.Team
	members     map[uint64]map[uint64]domain.TeamRole
	tasks       map[uint64]*domain.Task
	taskLogs    []*domain.TaskLog
	idempotency map[string]*domain.IdempotencyRecord

	notifications []domain.Notification
	notifyErr     error
}

// NewStore builds an empty Store.
func NewStore() *Store {
	return &Store{
		users:       map[uint64]*domain.User{},
		emailIndex:  map[string]uint64{},
		teams:       map[uint64]*domain.Team{},
		members:     map[uint64]map[uint64]domain.TeamRole{},
		tasks:       map[uint64]*domain.Task{},
		idempotency: map[string]*domain.IdempotencyRecord{},
	}
}

// SetNotifyError makes the notifier fail, for rollback tests.
func (s *Store) SetNotifyError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifyErr = err
}

// Notifications returns a copy of the recorded notifications.
func (s *Store) Notifications() []domain.Notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.Notification(nil), s.notifications...)
}

// TaskCount returns the number of tasks currently stored.
func (s *Store) TaskCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.tasks)
}

// TaskLogCount returns the number of audit logs currently stored.
func (s *Store) TaskLogCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.taskLogs)
}

// TaskAssignee returns the current assignee of a task, if any.
func (s *Store) TaskAssignee(taskID uint64) (*uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[taskID]
	if !ok || task.AssigneeID == nil {
		return nil, false
	}
	id := *task.AssigneeID
	return &id, true
}

// AddUser inserts a user directly, for test fixtures.
func (s *Store) AddUser(user domain.User) domain.User {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextUserID++
	user.ID = s.nextUserID
	now := time.Now()
	user.CreatedAt, user.UpdatedAt = now, now
	stored := user
	s.users[user.ID] = &stored
	s.emailIndex[user.Email] = user.ID
	return stored
}

// AddTeam inserts a team and registers the given member.
func (s *Store) AddTeam(name string, memberID uint64, role domain.TeamRole) domain.Team {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextTeamID++
	team := domain.Team{ID: s.nextTeamID, Name: name, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.teams[team.ID] = &team
	if s.members[team.ID] == nil {
		s.members[team.ID] = map[uint64]domain.TeamRole{}
	}
	s.members[team.ID][memberID] = role
	return team
}

// SetPrimaryTeam assigns a user's primary team, for test fixtures.
func (s *Store) SetPrimaryTeam(userID, teamID uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if user, ok := s.users[userID]; ok {
		id := teamID
		user.PrimaryTeamID = &id
	}
}

func (s *Store) snapshot() *Store {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cloneLocked()
}

func (s *Store) restore(from *Store) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextUserID = from.nextUserID
	s.nextTeamID = from.nextTeamID
	s.nextTaskID = from.nextTaskID
	s.users = from.users
	s.emailIndex = from.emailIndex
	s.teams = from.teams
	s.members = from.members
	s.tasks = from.tasks
	s.taskLogs = from.taskLogs
	s.idempotency = from.idempotency
	s.notifications = from.notifications
	s.notifyErr = from.notifyErr
}

func (s *Store) cloneLocked() *Store {
	clone := &Store{
		nextUserID:    s.nextUserID,
		nextTeamID:    s.nextTeamID,
		nextTaskID:    s.nextTaskID,
		users:         make(map[uint64]*domain.User, len(s.users)),
		emailIndex:    make(map[string]uint64, len(s.emailIndex)),
		teams:         make(map[uint64]*domain.Team, len(s.teams)),
		members:       make(map[uint64]map[uint64]domain.TeamRole, len(s.members)),
		tasks:         make(map[uint64]*domain.Task, len(s.tasks)),
		taskLogs:      append([]*domain.TaskLog(nil), s.taskLogs...),
		idempotency:   make(map[string]*domain.IdempotencyRecord, len(s.idempotency)),
		notifications: append([]domain.Notification(nil), s.notifications...),
		notifyErr:     s.notifyErr,
	}
	for id, user := range s.users {
		copied := *user
		clone.users[id] = &copied
	}
	for email, id := range s.emailIndex {
		clone.emailIndex[email] = id
	}
	for id, team := range s.teams {
		copied := *team
		clone.teams[id] = &copied
	}
	for teamID, members := range s.members {
		inner := make(map[uint64]domain.TeamRole, len(members))
		for userID, role := range members {
			inner[userID] = role
		}
		clone.members[teamID] = inner
	}
	for id, task := range s.tasks {
		copied := *task
		if task.AssigneeID != nil {
			assignee := *task.AssigneeID
			copied.AssigneeID = &assignee
		}
		clone.tasks[id] = &copied
	}
	for key, record := range s.idempotency {
		copied := *record
		copied.ResponseBody = append([]byte(nil), record.ResponseBody...)
		clone.idempotency[key] = &copied
	}
	return clone
}
