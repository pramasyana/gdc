package mocks

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/ryanpramasyana/gdc/test/domain"
)

// UserRepository is an in-memory domain.UserRepository.
type UserRepository struct {
	store *Store
}

// NewUserRepository builds an in-memory user repository.
func NewUserRepository(store *Store) *UserRepository { return &UserRepository{store: store} }

// Create implements domain.UserRepository.
func (r *UserRepository) Create(_ context.Context, user *domain.User) error {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.emailIndex[user.Email]; exists {
		return domain.ErrDuplicate
	}
	s.nextUserID++
	now := time.Now()
	user.ID = s.nextUserID
	user.CreatedAt, user.UpdatedAt = now, now
	stored := *user
	s.users[user.ID] = &stored
	s.emailIndex[user.Email] = user.ID
	return nil
}

// GetByID implements domain.UserRepository.
func (r *UserRepository) GetByID(_ context.Context, id uint64) (*domain.User, error) {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copied := *stored
	return &copied, nil
}

// GetByEmail implements domain.UserRepository.
func (r *UserRepository) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.emailIndex[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copied := *s.users[id]
	return &copied, nil
}

// Update implements domain.UserRepository.
func (r *UserRepository) Update(_ context.Context, user *domain.User) error {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.users[user.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if other, exists := s.emailIndex[user.Email]; exists && other != user.ID {
		return domain.ErrDuplicate
	}
	delete(s.emailIndex, stored.Email)
	stored.Name = user.Name
	stored.Email = user.Email
	stored.UpdatedAt = time.Now()
	s.emailIndex[user.Email] = user.ID
	return nil
}

// TeamRepository is an in-memory domain.TeamRepository.
type TeamRepository struct {
	store *Store
}

// NewTeamRepository builds an in-memory team repository.
func NewTeamRepository(store *Store) *TeamRepository { return &TeamRepository{store: store} }

// Create implements domain.TeamRepository.
func (r *TeamRepository) Create(_ context.Context, team *domain.Team) error {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextTeamID++
	now := time.Now()
	team.ID = s.nextTeamID
	team.CreatedAt, team.UpdatedAt = now, now
	stored := *team
	s.teams[team.ID] = &stored
	return nil
}

// AddMember implements domain.TeamRepository.
func (r *TeamRepository) AddMember(_ context.Context, teamID, userID uint64, role domain.TeamRole) error {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.members[teamID] == nil {
		s.members[teamID] = map[uint64]domain.TeamRole{}
	}
	s.members[teamID][userID] = role
	return nil
}

// IsMember implements domain.TeamRepository.
func (r *TeamRepository) IsMember(_ context.Context, teamID, userID uint64) (bool, error) {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.members[teamID][userID]
	return ok, nil
}

// TaskRepository is an in-memory domain.TaskRepository.
type TaskRepository struct {
	store *Store
}

// NewTaskRepository builds an in-memory task repository.
func NewTaskRepository(store *Store) *TaskRepository { return &TaskRepository{store: store} }

// Create implements domain.TaskRepository.
func (r *TaskRepository) Create(_ context.Context, task *domain.Task) error {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextTaskID++
	now := time.Now()
	task.ID = s.nextTaskID
	task.CreatedAt, task.UpdatedAt = now, now
	copied := *task
	s.tasks[task.ID] = &copied
	return nil
}

// GetByID implements domain.TaskRepository.
func (r *TaskRepository) GetByID(_ context.Context, id uint64) (*domain.Task, error) {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	return copyTask(s.tasks[id])
}

// List implements domain.TaskRepository.
func (r *TaskRepository) List(_ context.Context, filter domain.TaskFilter) (result []*domain.Task, total int64, err error) {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()

	var matched []*domain.Task
	for _, task := range s.tasks {
		if !matchesFilter(task, filter) {
			continue
		}
		copied := *task
		matched = append(matched, &copied)
	}
	total = int64(len(matched))
	start := filter.Offset()
	if start > len(matched) {
		return []*domain.Task{}, total, nil
	}
	end := start + filter.Limit
	if filter.Limit <= 0 || end > len(matched) {
		end = len(matched)
	}
	return matched[start:end], total, nil
}

// Update implements domain.TaskRepository.
func (r *TaskRepository) Update(_ context.Context, task *domain.Task) error {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.tasks[task.ID]
	if !ok {
		return domain.ErrNotFound
	}
	stored.Title = task.Title
	stored.Description = task.Description
	stored.Status = task.Status
	stored.DueDate = task.DueDate
	stored.UpdatedAt = time.Now()
	return nil
}

// UpdateAssignee implements domain.TaskRepository.
func (r *TaskRepository) UpdateAssignee(_ context.Context, taskID, assigneeID uint64) error {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.tasks[taskID]
	if !ok {
		return domain.ErrNotFound
	}
	id := assigneeID
	stored.AssigneeID = &id
	stored.UpdatedAt = time.Now()
	return nil
}

// Delete implements domain.TaskRepository.
func (r *TaskRepository) Delete(_ context.Context, id uint64) error {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return domain.ErrNotFound
	}
	delete(s.tasks, id)
	return nil
}

// TaskLogRepository is an in-memory domain.TaskLogRepository.
type TaskLogRepository struct {
	store *Store
}

// NewTaskLogRepository builds an in-memory task log repository.
func NewTaskLogRepository(store *Store) *TaskLogRepository { return &TaskLogRepository{store: store} }

// Create implements domain.TaskLogRepository.
func (r *TaskLogRepository) Create(_ context.Context, log *domain.TaskLog) error {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	log.ID = uint64(len(s.taskLogs)) + 1
	log.CreatedAt = time.Now()
	stored := *log
	s.taskLogs = append(s.taskLogs, &stored)
	return nil
}

// IdempotencyRepository is an in-memory domain.IdempotencyRepository whose
// Acquire is atomic, mirroring the unique constraint of the real store.
type IdempotencyRepository struct {
	store *Store
}

// NewIdempotencyRepository builds an in-memory idempotency repository.
func NewIdempotencyRepository(store *Store) *IdempotencyRepository {
	return &IdempotencyRepository{store: store}
}

// Acquire implements domain.IdempotencyRepository.
func (r *IdempotencyRepository) Acquire(_ context.Context, record *domain.IdempotencyRecord) (current *domain.IdempotencyRecord, acquired bool, err error) {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if existing, ok := s.idempotency[record.Key]; ok && !existing.IsExpired(now) {
		return copyRecord(existing), false, nil
	}
	stored := *record
	stored.CreatedAt = now
	stored.ResponseBody = append([]byte(nil), record.ResponseBody...)
	s.idempotency[record.Key] = &stored
	record.CreatedAt = now
	return nil, true, nil
}

// GetByKey implements domain.IdempotencyRepository.
func (r *IdempotencyRepository) GetByKey(_ context.Context, key string) (*domain.IdempotencyRecord, error) {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.idempotency[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return copyRecord(record), nil
}

// Complete implements domain.IdempotencyRepository.
func (r *IdempotencyRepository) Complete(_ context.Context, key string, status int, body []byte) error {
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.idempotency[key]
	if !ok {
		return domain.ErrNotFound
	}
	record.State = domain.IdempotencyStateCompleted
	record.ResponseStatus = status
	record.ResponseBody = append([]byte(nil), body...)
	return nil
}

// TxManager runs work as an in-memory transaction: it snapshots the store
// before running and restores it when the function fails.
type TxManager struct {
	store *Store
}

// NewTxManager builds an in-memory transaction manager.
func NewTxManager(store *Store) *TxManager { return &TxManager{store: store} }

// WithinTransaction implements domain.TxManager.
func (m *TxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	snapshot := m.store.snapshot()
	if err := fn(ctx); err != nil {
		m.store.restore(snapshot)
		return err
	}
	return nil
}

// Notifier records notifications in memory.
type Notifier struct {
	store *Store
}

// NewNotifier builds an in-memory notifier.
func NewNotifier(store *Store) *Notifier { return &Notifier{store: store} }

// Send implements domain.Notifier.
func (n *Notifier) Send(_ context.Context, notification domain.Notification) error {
	n.store.mu.Lock()
	defer n.store.mu.Unlock()
	if n.store.notifyErr != nil {
		return n.store.notifyErr
	}
	n.store.notifications = append(n.store.notifications, notification)
	return nil
}

// PasswordHasher is a trivial non-cryptographic hasher for tests.
type PasswordHasher struct{}

// NewPasswordHasher builds a test password hasher.
func NewPasswordHasher() *PasswordHasher { return &PasswordHasher{} }

// Hash implements domain.PasswordHasher.
func (h *PasswordHasher) Hash(plain string) (string, error) { return "hashed:" + plain, nil }

// Compare implements domain.PasswordHasher.
func (h *PasswordHasher) Compare(hash, plain string) error {
	if hash != "hashed:"+plain {
		return domain.ErrNotFound
	}
	return nil
}

// TokenProvider issues opaque, parseable tokens for tests.
type TokenProvider struct{}

// NewTokenProvider builds a test token provider.
func NewTokenProvider() *TokenProvider { return &TokenProvider{} }

// Generate implements domain.TokenProvider.
func (p *TokenProvider) Generate(userID uint64, tokenType domain.TokenType, ttl time.Duration) (claims *domain.TokenClaims, signed string, err error) {
	jti := uuid.NewString()
	claims = &domain.TokenClaims{UserID: userID, JTI: jti, Type: tokenType, Expiry: time.Now().Add(ttl)}
	signed = fmt.Sprintf("%s:%d:%s", tokenType, userID, jti)
	return claims, signed, nil
}

// Parse implements domain.TokenProvider.
func (p *TokenProvider) Parse(token string, tokenType domain.TokenType) (*domain.TokenClaims, error) {
	parts := strings.Split(token, ":")
	if len(parts) != 3 {
		return nil, domain.ErrNotFound
	}
	if parts[0] != string(tokenType) {
		return nil, domain.ErrNotFound
	}
	userID, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return &domain.TokenClaims{UserID: userID, JTI: parts[2], Type: tokenType}, nil
}

// TokenStore is an in-memory domain.TokenStore.
type TokenStore struct {
	mu        sync.Mutex
	refresh   map[string]uint64
	blacklist map[string]struct{}
}

// NewTokenStore builds an in-memory token store.
func NewTokenStore() *TokenStore {
	return &TokenStore{refresh: map[string]uint64{}, blacklist: map[string]struct{}{}}
}

// SaveRefresh implements domain.TokenStore.
func (s *TokenStore) SaveRefresh(_ context.Context, jti string, userID uint64, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refresh[jti] = userID
	return nil
}

// HasRefresh implements domain.TokenStore.
func (s *TokenStore) HasRefresh(_ context.Context, jti string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.refresh[jti]
	return ok, nil
}

// RevokeRefresh implements domain.TokenStore.
func (s *TokenStore) RevokeRefresh(_ context.Context, jti string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.refresh, jti)
	return nil
}

// BlacklistAccess implements domain.TokenStore.
func (s *TokenStore) BlacklistAccess(_ context.Context, jti string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blacklist[jti] = struct{}{}
	return nil
}

// IsAccessBlacklisted implements domain.TokenStore.
func (s *TokenStore) IsAccessBlacklisted(_ context.Context, jti string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.blacklist[jti]
	return ok, nil
}

func copyTask(task *domain.Task) (*domain.Task, error) {
	if task == nil {
		return nil, domain.ErrNotFound
	}
	copied := *task
	if task.AssigneeID != nil {
		assignee := *task.AssigneeID
		copied.AssigneeID = &assignee
	}
	return &copied, nil
}

func copyRecord(record *domain.IdempotencyRecord) *domain.IdempotencyRecord {
	copied := *record
	copied.ResponseBody = append([]byte(nil), record.ResponseBody...)
	return &copied
}

func matchesFilter(task *domain.Task, filter domain.TaskFilter) bool {
	if filter.OwnerID != nil {
		owner := *filter.OwnerID
		isOwner := task.OwnerID == owner
		isAssignee := task.AssigneeID != nil && *task.AssigneeID == owner
		if !isOwner && !isAssignee {
			return false
		}
	}
	if filter.Status != "" && task.Status != filter.Status {
		return false
	}
	if filter.Search != "" && !strings.Contains(strings.ToLower(task.Title), strings.ToLower(filter.Search)) {
		return false
	}
	return true
}
