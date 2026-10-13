// Package handler implements the auth module's HTTP handlers.
package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ryanpramasyana/gdc/internal/package/auth/handler/dto/request"
	authresp "github.com/ryanpramasyana/gdc/internal/package/auth/handler/dto/response"
	"github.com/ryanpramasyana/gdc/internal/package/auth/ports"
	"github.com/ryanpramasyana/gdc/internal/platform/middleware"
	"github.com/ryanpramasyana/gdc/internal/shared/httpx"
	"github.com/ryanpramasyana/gdc/internal/shared/response"
)

// Handler serves authentication endpoints.
type Handler struct {
	auth ports.AuthUsecase
}

// New builds an auth Handler.
func New(auth ports.AuthUsecase) *Handler {
	return &Handler{auth: auth}
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Creates a user with a personal team and returns a token pair.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request.RegisterRequest	true	"Registration payload"
//	@Success		201		{object}	response.SuccessResponse{data=authresp.AuthResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		409		{object}	response.ErrorResponse
//	@Router			/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := httpx.Bind(c, &req); err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.auth.Register(c.Request.Context(), ports.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		TeamName: req.TeamName,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "Created", authresp.FromAuthResult(result))
}

// Login godoc
//
//	@Summary		Log in
//	@Description	Authenticates a user and returns a token pair.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request.LoginRequest	true	"Login payload"
//	@Success		200		{object}	response.SuccessResponse{data=authresp.AuthResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := httpx.Bind(c, &req); err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.auth.Login(c.Request.Context(), ports.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "OK", authresp.FromAuthResult(result))
}

// Refresh godoc
//
//	@Summary		Refresh tokens
//	@Description	Rotates a valid refresh token and returns a new token pair.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request.RefreshRequest	true	"Refresh payload"
//	@Success		200		{object}	response.SuccessResponse{data=authresp.AuthResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req request.RefreshRequest
	if err := httpx.Bind(c, &req); err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "OK", authresp.FromAuthResult(result))
}

// Logout godoc
//
//	@Summary		Log out
//	@Description	Revokes the current access token and, when supplied, the refresh token.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		request.LogoutRequest	false	"Optional refresh token"
//	@Success		200		{object}	response.SuccessResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req request.LogoutRequest
	_ = c.ShouldBindJSON(&req)

	accessJTI := c.GetString(middleware.ContextAccessJTI)
	accessExp, _ := c.Get(middleware.ContextAccessExp)
	expiry, _ := accessExp.(time.Time)

	if err := h.auth.Logout(c.Request.Context(), userID, accessJTI, expiry, req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Logged out", nil)
}
