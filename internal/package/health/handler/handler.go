// Package handler implements the health module's liveness and readiness probes.
package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/ryanpramasyana/gdc/internal/platform/database"
	redisinfra "github.com/ryanpramasyana/gdc/internal/platform/redis"
	"github.com/ryanpramasyana/gdc/internal/shared/response"
)

// Handler serves liveness and readiness probes.
type Handler struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// New builds a health Handler.
func New(db *gorm.DB, redisClient *redis.Client) *Handler {
	return &Handler{db: db, redisClient: redisClient}
}

// Live godoc
//
//	@Summary	Liveness probe
//	@Tags		Health
//	@Produce	json
//	@Success	200	{object}	response.SuccessResponse
//	@Router		/health [get]
func (h *Handler) Live(c *gin.Context) {
	response.OK(c, "OK", gin.H{"status": "ok"})
}

// Ready godoc
//
//	@Summary	Readiness probe
//	@Description	Verifies connectivity to MySQL and Redis.
//	@Tags		Health
//	@Produce	json
//	@Success	200	{object}	response.SuccessResponse
//	@Failure	503	{object}	response.ErrorResponse
//	@Router		/ready [get]
func (h *Handler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := database.Ping(ctx, h.db); err != nil {
		response.Error(c, err)
		return
	}
	if err := redisinfra.Ping(ctx, h.redisClient); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "OK", gin.H{"status": "ready"})
}
