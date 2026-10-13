// Package handler implements the task module's HTTP handlers. Handlers only
// bind input, delegate to a usecase and render the response.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ryanpramasyana/gdc/internal/package/task/domain"
	"github.com/ryanpramasyana/gdc/internal/package/task/handler/dto/request"
	taskresp "github.com/ryanpramasyana/gdc/internal/package/task/handler/dto/response"
	"github.com/ryanpramasyana/gdc/internal/package/task/ports"
	"github.com/ryanpramasyana/gdc/internal/shared/httpx"
	"github.com/ryanpramasyana/gdc/internal/shared/response"
)

// Handler serves task management endpoints.
type Handler struct {
	tasks ports.TaskUsecase
}

// New builds a task Handler.
func New(tasks ports.TaskUsecase) *Handler {
	return &Handler{tasks: tasks}
}

// Create godoc
//
//	@Summary		Create a task
//	@Description	Creates a task. Requires an Idempotency-Key header; retries with the same key return the original response without creating a duplicate.
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Idempotency-Key	header		string					true	"UUID idempotency key"
//	@Param			request			body		request.CreateTaskRequest	true	"Task payload"
//	@Success		201				{object}	response.SuccessResponse{data=taskresp.TaskResponse}
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		409				{object}	response.ErrorResponse
//	@Router			/tasks [post]
func (h *Handler) Create(c *gin.Context) {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req request.CreateTaskRequest
	if err := httpx.Bind(c, &req); err != nil {
		response.Error(c, err)
		return
	}

	result, err := h.tasks.Create(
		c.Request.Context(),
		userID,
		c.GetHeader("Idempotency-Key"),
		ports.CreateTaskInput{
			Title:       req.Title,
			Description: req.Description,
			Status:      domain.TaskStatus(req.Status),
			DueDate:     req.DueDate,
		},
		renderCreatedTask,
	)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Raw(c, result.Status, result.Body)
}

// List godoc
//
//	@Summary		List tasks
//	@Description	Lists tasks the caller owns or is assigned to, with optional status filter, title search and pagination.
//	@Tags			Tasks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			status	query		string	false	"Filter by status"	Enums(todo, in_progress, done, cancelled)
//	@Param			search	query		string	false	"Search by title"
//	@Param			page	query		int		false	"Page number"			default(1)
//	@Param			limit	query		int		false	"Items per page"		default(10)
//	@Success		200		{object}	response.ListResponse{data=[]taskresp.TaskResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/tasks [get]
func (h *Handler) List(c *gin.Context) {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var query request.ListTasksQuery
	if err := httpx.BindQuery(c, &query); err != nil {
		response.Error(c, err)
		return
	}

	tasks, total, err := h.tasks.List(c.Request.Context(), userID, ports.ListTasksInput{
		Status: domain.TaskStatus(query.Status),
		Search: query.Search,
		Page:   query.Page,
		Limit:  query.Limit,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	limit := query.Limit
	if limit < 1 {
		limit = 10
	}
	response.List(c, "OK", taskresp.FromTasks(tasks), page, limit, total)
}

// Get godoc
//
//	@Summary		Get task detail
//	@Tags			Tasks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Task ID"
//	@Success		200	{object}	response.SuccessResponse{data=taskresp.TaskResponse}
//	@Failure		401	{object}	response.ErrorResponse
//	@Failure		403	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Router			/tasks/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	taskID, err := httpx.PathID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	task, err := h.tasks.Get(c.Request.Context(), userID, taskID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "OK", taskresp.FromTask(task))
}

// Update godoc
//
//	@Summary		Update a task
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"Task ID"
//	@Param			request	body		request.UpdateTaskRequest	true	"Task payload"
//	@Success		200		{object}	response.SuccessResponse{data=taskresp.TaskResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Failure		403		{object}	response.ErrorResponse
//	@Failure		404		{object}	response.ErrorResponse
//	@Router			/tasks/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	taskID, err := httpx.PathID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req request.UpdateTaskRequest
	if err := httpx.Bind(c, &req); err != nil {
		response.Error(c, err)
		return
	}
	task, err := h.tasks.Update(c.Request.Context(), userID, taskID, ports.UpdateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      domain.TaskStatus(req.Status),
		DueDate:     req.DueDate,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "OK", taskresp.FromTask(task))
}

// Delete godoc
//
//	@Summary		Delete a task
//	@Tags			Tasks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Task ID"
//	@Success		200	{object}	response.SuccessResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Failure		403	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Router			/tasks/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	taskID, err := httpx.PathID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.tasks.Delete(c.Request.Context(), userID, taskID); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Task deleted", nil)
}

// Assign godoc
//
//	@Summary		Assign a task
//	@Description	Assigns the task to another member of the task's team inside a single transaction that also writes the audit log and dispatches a notification.
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"Task ID"
//	@Param			request	body		request.AssignTaskRequest	true	"Assignee payload"
//	@Success		200		{object}	response.SuccessResponse{data=taskresp.TaskResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Failure		403		{object}	response.ErrorResponse
//	@Failure		404		{object}	response.ErrorResponse
//	@Failure		422		{object}	response.ErrorResponse
//	@Router			/tasks/{id}/assign [post]
func (h *Handler) Assign(c *gin.Context) {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	taskID, err := httpx.PathID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req request.AssignTaskRequest
	if err := httpx.Bind(c, &req); err != nil {
		response.Error(c, err)
		return
	}
	task, err := h.tasks.Assign(c.Request.Context(), userID, taskID, req.AssigneeID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "OK", taskresp.FromTask(task))
}

// renderCreatedTask serializes a freshly created task into the exact bytes
// stored for idempotent replays.
func renderCreatedTask(task *domain.Task) (status int, body []byte) {
	body, err := json.Marshal(response.NewSuccess("Created", taskresp.FromTask(task)))
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	return http.StatusCreated, body
}
