// Package handler implements the user module's HTTP handlers.
package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/ryanpramasyana/gdc/internal/package/user/handler/dto/request"
	userresp "github.com/ryanpramasyana/gdc/internal/package/user/handler/dto/response"
	"github.com/ryanpramasyana/gdc/internal/package/user/ports"
	"github.com/ryanpramasyana/gdc/internal/shared/httpx"
	"github.com/ryanpramasyana/gdc/internal/shared/response"
)

// Handler serves user profile endpoints.
type Handler struct {
	users ports.UserUsecase
}

// New builds a user Handler.
func New(users ports.UserUsecase) *Handler {
	return &Handler{users: users}
}

// Me godoc
//
//	@Summary	Get current user
//	@Tags		Users
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	response.SuccessResponse{data=userresp.UserResponse}
//	@Failure	401	{object}	response.ErrorResponse
//	@Router		/users/me [get]
func (h *Handler) Me(c *gin.Context) {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	user, err := h.users.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "OK", userresp.FromUser(user))
}

// UpdateMe godoc
//
//	@Summary	Update current user
//	@Tags		Users
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		request.UpdateProfileRequest	true	"Profile payload"
//	@Success	200		{object}	response.SuccessResponse{data=userresp.UserResponse}
//	@Failure	400		{object}	response.ErrorResponse
//	@Failure	401		{object}	response.ErrorResponse
//	@Failure	409		{object}	response.ErrorResponse
//	@Router		/users/me [put]
func (h *Handler) UpdateMe(c *gin.Context) {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req request.UpdateProfileRequest
	if err := httpx.Bind(c, &req); err != nil {
		response.Error(c, err)
		return
	}
	user, err := h.users.UpdateProfile(c.Request.Context(), userID, ports.UpdateProfileInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "OK", userresp.FromUser(user))
}
