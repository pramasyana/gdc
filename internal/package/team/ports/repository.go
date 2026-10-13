package ports

import (
	"context"

	"github.com/ryanpramasyana/gdc/internal/package/team/domain"
)

// TeamRepository persists teams and their memberships.
type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	AddMember(ctx context.Context, teamID, userID uint64, role domain.TeamRole) error
	IsMember(ctx context.Context, teamID, userID uint64) (bool, error)
}
