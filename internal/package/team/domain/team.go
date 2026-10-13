// Package domain holds the team module's entities. It imports no framework.
package domain

import "time"

// Team is a group of users. Tasks belong to the primary team of their owner.
type Team struct {
	ID        uint64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TeamRole describes a member's role within a team.
type TeamRole string

// Supported team roles.
const (
	RoleOwner  TeamRole = "owner"
	RoleMember TeamRole = "member"
)
