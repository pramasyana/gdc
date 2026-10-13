package main

import (
	"context"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	taskdomain "github.com/ryanpramasyana/gdc/internal/package/task/domain"
	taskpkg "github.com/ryanpramasyana/gdc/internal/package/task/repository"
	"github.com/ryanpramasyana/gdc/internal/package/team/domain"
	teampkg "github.com/ryanpramasyana/gdc/internal/package/team/repository"
	userdomain "github.com/ryanpramasyana/gdc/internal/package/user/domain"
	userpkg "github.com/ryanpramasyana/gdc/internal/package/user/repository"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "gdc:gdc_secret@tcp(localhost:3306)/gdc?parseTime=true&charset=utf8mb4"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect db:", err)
	}

	teams := teampkg.New(db)
	users := userpkg.New(db)
	tasks := taskpkg.New(db)

	ctx := context.Background()

	team := &domain.Team{Name: "Demo Team"}
	if err := teams.Create(ctx, team); err != nil {
		log.Fatal(err)
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	u1 := &userdomain.User{Name: "Demo User", Email: "demo@example.com", PasswordHash: string(hash), PrimaryTeamID: &team.ID}
	if err := users.Create(ctx, u1); err != nil {
		log.Fatal(err)
	}
	if err := teams.AddMember(ctx, team.ID, u1.ID, domain.RoleOwner); err != nil {
		log.Fatal(err)
	}

	u2 := &userdomain.User{Name: "Alice", Email: "alice@example.com", PasswordHash: string(hash), PrimaryTeamID: &team.ID}
	if err := users.Create(ctx, u2); err != nil {
		log.Fatal(err)
	}
	if err := teams.AddMember(ctx, team.ID, u2.ID, domain.RoleMember); err != nil {
		log.Fatal(err)
	}

	t1 := &taskdomain.Task{Title: "Write docs", Description: "README & API", Status: taskdomain.TaskStatusTodo, OwnerID: u1.ID, TeamID: team.ID}
	t2 := &taskdomain.Task{Title: "Fix bug", Description: "Idem test", Status: taskdomain.TaskStatusInProgress, OwnerID: u1.ID, TeamID: team.ID}
	t3 := &taskdomain.Task{Title: "Review PR", Status: taskdomain.TaskStatusTodo, OwnerID: u2.ID, TeamID: team.ID}
	_ = tasks.Create(ctx, t1)
	_ = tasks.Create(ctx, t2)
	_ = tasks.Create(ctx, t3)

	log.Println("seed completed: demo@example.com/password123, alice@example.com/password123")
}
