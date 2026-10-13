// Package domain aggregates the per-module domain entities and the shared
// kernel contracts under a single namespace so the in-memory test fakes can
// reference them without importing every module individually. It exists only to
// support tests and must never be imported by production code.
package domain

import (
	idemdomain "github.com/ryanpramasyana/gdc/internal/package/idempotency/domain"
	taskdomain "github.com/ryanpramasyana/gdc/internal/package/task/domain"
	teamdomain "github.com/ryanpramasyana/gdc/internal/package/team/domain"
	userdomain "github.com/ryanpramasyana/gdc/internal/package/user/domain"
	"github.com/ryanpramasyana/gdc/internal/shared/kernel"
)

type (
	User              = userdomain.User
	Team              = teamdomain.Team
	TeamRole          = teamdomain.TeamRole
	Task              = taskdomain.Task
	TaskStatus        = taskdomain.TaskStatus
	TaskFilter        = taskdomain.TaskFilter
	TaskLog           = taskdomain.TaskLog
	IdempotencyRecord = idemdomain.IdempotencyRecord
	IdempotencyState  = idemdomain.IdempotencyState

	Notification = kernel.Notification
	Notifier     = kernel.Notifier
	TokenType    = kernel.TokenType
	TokenClaims  = kernel.TokenClaims
)

const (
	RoleOwner  = teamdomain.RoleOwner
	RoleMember = teamdomain.RoleMember

	TaskStatusTodo       = taskdomain.TaskStatusTodo
	TaskStatusInProgress = taskdomain.TaskStatusInProgress
	TaskStatusDone       = taskdomain.TaskStatusDone
	TaskStatusCancelled  = taskdomain.TaskStatusCancelled

	ActionAssigned = taskdomain.ActionAssigned

	IdempotencyStateProcessing = idemdomain.IdempotencyStateProcessing
	IdempotencyStateCompleted  = idemdomain.IdempotencyStateCompleted

	TokenTypeAccess  = kernel.TokenTypeAccess
	TokenTypeRefresh = kernel.TokenTypeRefresh
)

var (
	ErrNotFound  = kernel.ErrNotFound
	ErrDuplicate = kernel.ErrDuplicate
)
