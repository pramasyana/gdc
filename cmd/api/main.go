// Package main is the application entrypoint. It loads configuration, wires
// all dependencies and starts the HTTP server with graceful shutdown.
//
//	@title						GDC Task Management API
//	@version					1.0
//	@description				Multi-user task management API with JWT auth, idempotent task creation and transactional assignment.
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@example.com
//	@host						localhost:8080
//	@BasePath					/
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and the JWT access token.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	httpdelivery "github.com/ryanpramasyana/gdc/internal/http"
	authhandler "github.com/ryanpramasyana/gdc/internal/package/auth/handler"
	authusecase "github.com/ryanpramasyana/gdc/internal/package/auth/usecase"
	healthhandler "github.com/ryanpramasyana/gdc/internal/package/health/handler"
	idemrepo "github.com/ryanpramasyana/gdc/internal/package/idempotency/repository"
	taskdomain "github.com/ryanpramasyana/gdc/internal/package/task/domain"
	taskhandler "github.com/ryanpramasyana/gdc/internal/package/task/handler"
	taskpkg "github.com/ryanpramasyana/gdc/internal/package/task/repository"
	taskusecase "github.com/ryanpramasyana/gdc/internal/package/task/usecase"
	teamdomain "github.com/ryanpramasyana/gdc/internal/package/team/domain"
	teampkg "github.com/ryanpramasyana/gdc/internal/package/team/repository"
	userdomain "github.com/ryanpramasyana/gdc/internal/package/user/domain"
	userhandler "github.com/ryanpramasyana/gdc/internal/package/user/handler"
	userpkg "github.com/ryanpramasyana/gdc/internal/package/user/repository"
	userusecase "github.com/ryanpramasyana/gdc/internal/package/user/usecase"
	"github.com/ryanpramasyana/gdc/internal/platform/cache"
	"github.com/ryanpramasyana/gdc/internal/platform/config"
	"github.com/ryanpramasyana/gdc/internal/platform/database"
	"github.com/ryanpramasyana/gdc/internal/platform/logger"
	"github.com/ryanpramasyana/gdc/internal/platform/notify"
	redisinfra "github.com/ryanpramasyana/gdc/internal/platform/redis"
	"github.com/ryanpramasyana/gdc/internal/platform/security"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func seedData(db *gorm.DB) error {
	var count int64
	if err := db.Model(&userdomain.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	teams := teampkg.New(db)
	users := userpkg.New(db)
	tasks := taskpkg.New(db)

	ctx := context.Background()

	team := &teamdomain.Team{Name: "Demo Team"}
	if err := teams.Create(ctx, team); err != nil {
		return err
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	u1 := &userdomain.User{Name: "Demo User", Email: "demo@example.com", PasswordHash: string(hash), PrimaryTeamID: &team.ID}
	if err := users.Create(ctx, u1); err != nil {
		return err
	}
	if err := teams.AddMember(ctx, team.ID, u1.ID, teamdomain.RoleOwner); err != nil {
		return err
	}

	u2 := &userdomain.User{Name: "Alice", Email: "alice@example.com", PasswordHash: string(hash), PrimaryTeamID: &team.ID}
	if err := users.Create(ctx, u2); err != nil {
		return err
	}
	if err := teams.AddMember(ctx, team.ID, u2.ID, teamdomain.RoleMember); err != nil {
		return err
	}

	t1 := &taskdomain.Task{Title: "Write docs", Description: "README & API", Status: taskdomain.TaskStatusTodo, OwnerID: u1.ID, TeamID: team.ID}
	t2 := &taskdomain.Task{Title: "Fix bug", Description: "Idem test", Status: taskdomain.TaskStatusInProgress, OwnerID: u1.ID, TeamID: team.ID}
	t3 := &taskdomain.Task{Title: "Review PR", Status: taskdomain.TaskStatusTodo, OwnerID: u2.ID, TeamID: team.ID}
	if err := tasks.Create(ctx, t1); err != nil {
		return err
	}
	if err := tasks.Create(ctx, t2); err != nil {
		return err
	}
	if err := tasks.Create(ctx, t3); err != nil {
		return err
	}
	slog.Info("seeded dummy data", "users", 2)
	return nil
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config load: %w", err)
	}

	appLogger := logger.New("info")
	slog.SetDefault(appLogger)

	redisClient := redisinfra.New(cfg.Redis)
	if err != nil {
		return fmt.Errorf("redis connect: %w", err)
	}
	defer redisClient.Close()

	db, err := database.New(cfg.DB, cfg.App.Env)
	if err != nil {
		return fmt.Errorf("database connect: %w", err)
	}

	if err := seedData(db); err != nil {
		return fmt.Errorf("seed data: %w", err)
	}

	hasher := security.NewPasswordHasher()
	tokenProv := security.NewJWTProvider(cfg.JWT.Secret, cfg.JWT.Issuer)
	tokenStore := cache.NewTokenStore(redisClient)

	tasks := taskpkg.New(db)
	taskLogs := taskpkg.NewTaskLog(db)
	users := userpkg.New(db)
	teams := teampkg.New(db)
	idempotency := idemrepo.New(db)
	txManager := database.NewTxManager(db)
	notifier := notify.NewLogNotifier(appLogger)

	authUC := authusecase.New(users, teams, hasher, tokenProv, tokenStore, txManager, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	taskUC := taskusecase.New(tasks, taskLogs, users, teams, idempotency, notifier, txManager, cfg.IdempotencyTTL())
	userUC := userusecase.New(users)

	authHandler := authhandler.New(authUC)
	taskHandler := taskhandler.New(taskUC)
	userHandler := userhandler.New(userUC)
	healthHandler := healthhandler.New(db, redisClient)

	var swaggerHandler gin.HandlerFunc
	if !cfg.App.IsProduction() {
		swaggerHandler = ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json"))
	}

	router := httpdelivery.NewRouter(httpdelivery.Deps{
		Config:        cfg,
		Logger:        appLogger,
		Tokens:        tokenProv,
		TokenStore:    tokenStore,
		AuthHandler:   authHandler,
		TaskHandler:   taskHandler,
		UserHandler:   userHandler,
		HealthHandler: healthHandler,
		Swagger:       swaggerHandler,
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		appLogger.Info("http server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("http server failed", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	return nil
}
