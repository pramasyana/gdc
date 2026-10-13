//go:build integration

// Package integration contains tests that exercise the real MySQL repositories.
// They require a running MySQL (and Redis) instance with migrations applied:
//
//	make docker-up && make migrate-up
//	go test -tags integration ./test/integration/...
package integration

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	idemrepo "github.com/ryanpramasyana/gdc/internal/package/idempotency/repository"
	taskports "github.com/ryanpramasyana/gdc/internal/package/task/ports"
	taskrepo "github.com/ryanpramasyana/gdc/internal/package/task/repository"
	taskusecase "github.com/ryanpramasyana/gdc/internal/package/task/usecase"
	teamrepo "github.com/ryanpramasyana/gdc/internal/package/team/repository"
	userrepo "github.com/ryanpramasyana/gdc/internal/package/user/repository"
	"github.com/ryanpramasyana/gdc/internal/platform/config"
	"github.com/ryanpramasyana/gdc/internal/platform/database"
	"github.com/ryanpramasyana/gdc/internal/platform/notify"
	"github.com/ryanpramasyana/gdc/test/domain"
)

// TestIdempotencyConcurrentRealMySQL proves the database-level guarantee: many
// concurrent requests with the same Idempotency-Key create exactly one task.
func TestIdempotencyConcurrentRealMySQL(t *testing.T) {
	db := openDB(t)
	cleanup(t, db)

	users := userrepo.New(db)
	teams := teamrepo.New(db)
	tasks := taskrepo.New(db)
	taskLogs := taskrepo.NewTaskLog(db)
	idempotency := idemrepo.New(db)
	txManager := database.NewTxManager(db)
	notifier := notify.NewLogNotifier(slog.New(slog.NewTextHandler(io.Discard, nil)))

	ctx := context.Background()

	team := &domain.Team{Name: "integration"}
	require.NoError(t, teams.Create(ctx, team))
	user := &domain.User{Name: "Owner", Email: uuid.NewString() + "@example.com", PasswordHash: "x", PrimaryTeamID: &team.ID}
	require.NoError(t, users.Create(ctx, user))
	require.NoError(t, teams.AddMember(ctx, team.ID, user.ID, domain.RoleOwner))

	taskUsecase := taskusecase.New(tasks, taskLogs, users, teams, idempotency, notifier, txManager, 24*time.Hour)

	const goroutines = 32
	key := uuid.NewString()
	input := taskports.CreateTaskInput{Title: "Concurrent integration"}
	render := func(task *domain.Task) (int, []byte) {
		return 201, []byte(`{"id":` + itoa(task.ID) + `}`)
	}

	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make([]error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			_, err := taskUsecase.Create(ctx, user.ID, key, input, render)
			errs[index] = err
		}(i)
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "goroutine %d", i)
	}

	var count int64
	require.NoError(t, db.Model(&struct{}{}).Table("tasks").Where("owner_id = ?", user.ID).Count(&count).Error)
	require.Equal(t, int64(1), count, "exactly one task must be persisted")
}

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	cfg := config.DB{
		Host:            envOr("DB_HOST", "localhost"),
		Port:            envOr("DB_PORT", "3306"),
		User:            envOr("DB_USER", "gdc"),
		Password:        envOr("DB_PASSWORD", "gdc_secret"),
		Name:            envOr("DB_NAME", "gdc"),
		MaxOpenConns:    goroutinesPool,
		MaxIdleConns:    goroutinesPool,
		ConnMaxLifetime: 5 * time.Minute,
	}
	db, err := database.New(cfg, "development")
	require.NoError(t, err, "start MySQL with `make docker-up` and apply migrations with `make migrate-up`")
	return db
}

const goroutinesPool = 40

func cleanup(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec("DELETE FROM idempotency_keys").Error)
	require.NoError(t, db.Exec("DELETE FROM task_logs").Error)
	require.NoError(t, db.Exec("DELETE FROM tasks").Error)
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func itoa(value uint64) string {
	if value == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for value > 0 {
		pos--
		buf[pos] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[pos:])
}
