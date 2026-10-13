// Package domain holds the user module's entities. It imports no framework.
package domain

import "time"

// User is an application user. PasswordHash is never serialized.
type User struct {
	ID            uint64
	Name          string
	Email         string
	PasswordHash  string
	PrimaryTeamID *uint64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
