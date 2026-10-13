// Package http is the delivery composition root: it builds the middleware chain
// and routes for every feature module.
package http

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	authhandler "github.com/ryanpramasyana/gdc/internal/package/auth/handler"
	healthhandler "github.com/ryanpramasyana/gdc/internal/package/health/handler"
	taskhandler "github.com/ryanpramasyana/gdc/internal/package/task/handler"
	userhandler "github.com/ryanpramasyana/gdc/internal/package/user/handler"
	"github.com/ryanpramasyana/gdc/internal/platform/config"
	"github.com/ryanpramasyana/gdc/internal/platform/middleware"
	"github.com/ryanpramasyana/gdc/internal/shared/kernel"
)

// Deps bundles the dependencies required to build the router.
type Deps struct {
	Config        *config.Config
	Logger        *slog.Logger
	Tokens        kernel.TokenProvider
	TokenStore    kernel.TokenStore
	AuthHandler   *authhandler.Handler
	TaskHandler   *taskhandler.Handler
	UserHandler   *userhandler.Handler
	HealthHandler *healthhandler.Handler
	Swagger       gin.HandlerFunc
}

// NewRouter builds the Gin engine with the global middleware chain and routes.
func NewRouter(deps Deps) *gin.Engine {
	if deps.Config.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(
		middleware.RequestID(),
		middleware.Recovery(deps.Logger),
		middleware.Logger(deps.Logger),
		middleware.CORS(),
	)

	engine.GET("/health", deps.HealthHandler.Live)
	engine.GET("/ready", deps.HealthHandler.Ready)

	if deps.Swagger != nil {
		engine.GET("/swagger/*any", deps.Swagger)
	}

	registerAuthRoutes(engine, deps)
	registerUserRoutes(engine, deps)
	registerTaskRoutes(engine, deps)

	return engine
}

func registerAuthRoutes(engine *gin.Engine, deps Deps) {
	auth := engine.Group("/auth")
	auth.POST("/register", deps.AuthHandler.Register)
	auth.POST("/login", deps.AuthHandler.Login)
	auth.POST("/refresh", deps.AuthHandler.Refresh)
	auth.POST("/logout", middleware.Auth(deps.Tokens, deps.TokenStore), deps.AuthHandler.Logout)
}

func registerUserRoutes(engine *gin.Engine, deps Deps) {
	users := engine.Group("/users", middleware.Auth(deps.Tokens, deps.TokenStore))
	users.GET("/me", deps.UserHandler.Me)
	users.PUT("/me", deps.UserHandler.UpdateMe)
}

func registerTaskRoutes(engine *gin.Engine, deps Deps) {
	tasks := engine.Group("/tasks", middleware.Auth(deps.Tokens, deps.TokenStore))
	tasks.POST("", deps.TaskHandler.Create)
	tasks.GET("", deps.TaskHandler.List)
	tasks.GET("/:id", deps.TaskHandler.Get)
	tasks.PUT("/:id", deps.TaskHandler.Update)
	tasks.DELETE("/:id", deps.TaskHandler.Delete)
	tasks.POST("/:id/assign", deps.TaskHandler.Assign)
}
