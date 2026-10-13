// Package usecase implements the task module's business rules. It depends only
// on the module's domain and ports, the user/team/idempotency ports it
// collaborates with, and the shared kernel/errors.
package usecase

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/ryanpramasyana/gdc/internal/shared/apperr"
	"github.com/ryanpramasyana/gdc/internal/shared/kernel"

	idemdomain "github.com/ryanpramasyana/gdc/internal/package/idempotency/domain"
	idemports "github.com/ryanpramasyana/gdc/internal/package/idempotency/ports"
	"github.com/ryanpramasyana/gdc/internal/package/task/domain"
	"github.com/ryanpramasyana/gdc/internal/package/task/ports"
	teamports "github.com/ryanpramasyana/gdc/internal/package/team/ports"
	userports "github.com/ryanpramasyana/gdc/internal/package/user/ports"
)

const (
	defaultPage       = 1
	defaultLimit      = 10
	maxLimit          = 100
	completionTimeout = 2 * time.Second
	completionPoll    = 10 * time.Millisecond
)

type taskUsecase struct {
	tasks          ports.TaskRepository
	taskLogs       ports.TaskLogRepository
	users          userports.UserRepository
	teams          teamports.TeamRepository
	idempotency    idemports.IdempotencyRepository
	notifier       kernel.Notifier
	tx             kernel.TxManager
	idempotencyTTL time.Duration
}

// New builds the task management usecase.
func New(
	tasks ports.TaskRepository,
	taskLogs ports.TaskLogRepository,
	users userports.UserRepository,
	teams teamports.TeamRepository,
	idempotency idemports.IdempotencyRepository,
	notifier kernel.Notifier,
	tx kernel.TxManager,
	idempotencyTTL time.Duration,
) ports.TaskUsecase {
	return &taskUsecase{
		tasks:          tasks,
		taskLogs:       taskLogs,
		users:          users,
		teams:          teams,
		idempotency:    idempotency,
		notifier:       notifier,
		tx:             tx,
		idempotencyTTL: idempotencyTTL,
	}
}

// Create creates a task guarded by an idempotency key. The key's uniqueness
// constraint, the task insert and the stored response all happen in one
// transaction, so concurrent duplicates create exactly one task.
func (u *taskUsecase) Create(
	ctx context.Context,
	userID uint64,
	idempotencyKey string,
	input ports.CreateTaskInput,
	render ports.RenderFunc,
) (*ports.IdempotentResult, error) {
	if err := validateIdempotencyKey(idempotencyKey); err != nil {
		return nil, err
	}
	if err := validateTitle(input.Title); err != nil {
		return nil, err
	}
	status, err := normalizeStatus(input.Status)
	if err != nil {
		return nil, err
	}

	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		return nil, apperr.TranslateNotFound(err, "user")
	}
	if user.PrimaryTeamID == nil {
		return nil, apperr.Internal(errors.New("user has no primary team"))
	}

	requestHash := hashCreateRequest(userID, input)

	var result *ports.IdempotentResult
	var inProgress bool
	err = u.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		now := time.Now()
		record := &idemdomain.IdempotencyRecord{
			Key:         idempotencyKey,
			UserID:      userID,
			Method:      "POST",
			Path:        "/tasks",
			RequestHash: requestHash,
			State:       idemdomain.IdempotencyStateProcessing,
			CreatedAt:   now,
			ExpiresAt:   now.Add(u.idempotencyTTL),
		}

		existing, acquired, err := u.idempotency.Acquire(ctx, record)
		if err != nil {
			return err
		}
		if !acquired {
			if existing.RequestHash != requestHash {
				return apperr.Conflict(apperr.CodeIdempotencyMismatch, "Idempotency-Key was already used with a different request payload")
			}
			if existing.IsCompleted() {
				result = &ports.IdempotentResult{Status: existing.ResponseStatus, Body: existing.ResponseBody, Replayed: true}
				return nil
			}
			inProgress = true
			return nil
		}

		task := &domain.Task{
			OwnerID:     userID,
			TeamID:      *user.PrimaryTeamID,
			Title:       strings.TrimSpace(input.Title),
			Description: input.Description,
			Status:      status,
			DueDate:     input.DueDate,
		}
		if err := u.tasks.Create(ctx, task); err != nil {
			return err
		}
		responseStatus, body := render(task)
		if err := u.idempotency.Complete(ctx, idempotencyKey, responseStatus, body); err != nil {
			return err
		}
		result = &ports.IdempotentResult{Status: responseStatus, Body: body}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if inProgress {
		return u.resolveInFlight(ctx, idempotencyKey, requestHash)
	}
	return result, nil
}

// List returns tasks the user owns or is assigned to, filtered and paginated.
func (u *taskUsecase) List(ctx context.Context, userID uint64, input ports.ListTasksInput) ([]*domain.Task, int64, error) {
	status := input.Status
	if status != "" && !status.Valid() {
		return nil, 0, apperr.BadRequest(apperr.CodeValidation, "invalid task status")
	}
	filter := domain.TaskFilter{
		OwnerID: &userID,
		Status:  status,
		Search:  strings.TrimSpace(input.Search),
		Pagination: kernel.Pagination{
			Page:  normalizePage(input.Page),
			Limit: normalizeLimit(input.Limit),
		},
	}
	return u.tasks.List(ctx, filter)
}

// Get returns a task the user may access.
func (u *taskUsecase) Get(ctx context.Context, userID, taskID uint64) (*domain.Task, error) {
	task, err := u.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, apperr.TranslateNotFound(err, "task")
	}
	if !canAccess(userID, task) {
		return nil, apperr.Forbidden("you do not have access to this task")
	}
	return task, nil
}

// Update modifies a task owned by the user.
func (u *taskUsecase) Update(ctx context.Context, userID, taskID uint64, input ports.UpdateTaskInput) (*domain.Task, error) {
	if err := validateTitle(input.Title); err != nil {
		return nil, err
	}
	status, err := normalizeStatus(input.Status)
	if err != nil {
		return nil, err
	}
	task, err := u.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, apperr.TranslateNotFound(err, "task")
	}
	if task.OwnerID != userID {
		return nil, apperr.Forbidden("only the task owner can update it")
	}
	task.Title = strings.TrimSpace(input.Title)
	task.Description = input.Description
	task.Status = status
	task.DueDate = input.DueDate
	if err := u.tasks.Update(ctx, task); err != nil {
		return nil, apperr.TranslateNotFound(err, "task")
	}
	return task, nil
}

// Delete soft-deletes a task owned by the user.
func (u *taskUsecase) Delete(ctx context.Context, userID, taskID uint64) error {
	task, err := u.tasks.GetByID(ctx, taskID)
	if err != nil {
		return apperr.TranslateNotFound(err, "task")
	}
	if task.OwnerID != userID {
		return apperr.Forbidden("only the task owner can delete it")
	}
	return apperr.TranslateNotFound(u.tasks.Delete(ctx, taskID), "task")
}

// Assign assigns a task to another member of the task's team. The assignee
// update, the audit log and the notification happen in one transaction.
func (u *taskUsecase) Assign(ctx context.Context, actorID, taskID, assigneeID uint64) (*domain.Task, error) {
	var updated *domain.Task
	err := u.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		task, err := u.tasks.GetByID(ctx, taskID)
		if err != nil {
			return apperr.TranslateNotFound(err, "task")
		}
		if task.OwnerID != actorID {
			return apperr.Forbidden("only the task owner can assign it")
		}
		if assigneeID == actorID {
			return apperr.Unprocessable("task is already owned by this user")
		}

		member, err := u.teams.IsMember(ctx, task.TeamID, assigneeID)
		if err != nil {
			return err
		}
		if !member {
			return apperr.Unprocessable("assignee is not a member of the task's team")
		}
		if _, err := u.users.GetByID(ctx, assigneeID); err != nil {
			return apperr.TranslateNotFound(err, "assignee")
		}

		oldValue := ""
		if task.AssigneeID != nil {
			oldValue = strconv.FormatUint(*task.AssigneeID, 10)
		}
		if err := u.tasks.UpdateAssignee(ctx, taskID, assigneeID); err != nil {
			return apperr.TranslateNotFound(err, "task")
		}
		log := &domain.TaskLog{
			TaskID:   taskID,
			ActorID:  actorID,
			Action:   domain.ActionAssigned,
			OldValue: jsonString(oldValue),
			NewValue: jsonString(strconv.FormatUint(assigneeID, 10)),
		}
		if err := u.taskLogs.Create(ctx, log); err != nil {
			return err
		}
		if err := u.notifier.Send(ctx, kernel.Notification{
			UserID:   assigneeID,
			Type:     "task_assigned",
			Title:    "A task was assigned to you",
			Body:     task.Title,
			Metadata: map[string]any{"task_id": taskID, "assigned_by": actorID},
		}); err != nil {
			return err
		}
		task.AssigneeID = &assigneeID
		updated = task
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// resolveInFlight waits briefly for a concurrent request holding the same key
// to finish, then replays its stored response.
func (u *taskUsecase) resolveInFlight(ctx context.Context, key, requestHash string) (*ports.IdempotentResult, error) {
	deadline := time.Now().Add(completionTimeout)
	for time.Now().Before(deadline) {
		record, err := u.idempotency.GetByKey(ctx, key)
		if err == nil {
			if record.RequestHash != requestHash {
				return nil, apperr.Conflict(apperr.CodeIdempotencyMismatch, "Idempotency-Key was already used with a different request payload")
			}
			if record.IsCompleted() {
				return &ports.IdempotentResult{Status: record.ResponseStatus, Body: record.ResponseBody, Replayed: true}, nil
			}
		} else if !errors.Is(err, kernel.ErrNotFound) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(completionPoll):
		}
	}
	return nil, apperr.Conflict(apperr.CodeIdempotencyInProgress, "a request with the same Idempotency-Key is being processed")
}
