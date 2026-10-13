// Package usecase implements the authentication business rules. It depends only
// on the auth ports and the user/team ports it collaborates with, plus the
// shared kernel/errors.
package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ryanpramasyana/gdc/internal/package/auth/ports"
	teamdomain "github.com/ryanpramasyana/gdc/internal/package/team/domain"
	teamports "github.com/ryanpramasyana/gdc/internal/package/team/ports"
	userdomain "github.com/ryanpramasyana/gdc/internal/package/user/domain"
	userports "github.com/ryanpramasyana/gdc/internal/package/user/ports"
	"github.com/ryanpramasyana/gdc/internal/shared/apperr"
	"github.com/ryanpramasyana/gdc/internal/shared/kernel"
)

type authUsecase struct {
	users      userports.UserRepository
	teams      teamports.TeamRepository
	hasher     kernel.PasswordHasher
	tokens     kernel.TokenProvider
	tokenStore kernel.TokenStore
	tx         kernel.TxManager
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// New builds the authentication usecase.
func New(
	users userports.UserRepository,
	teams teamports.TeamRepository,
	hasher kernel.PasswordHasher,
	tokens kernel.TokenProvider,
	tokenStore kernel.TokenStore,
	tx kernel.TxManager,
	accessTTL, refreshTTL time.Duration,
) ports.AuthUsecase {
	return &authUsecase{
		users:      users,
		teams:      teams,
		hasher:     hasher,
		tokens:     tokens,
		tokenStore: tokenStore,
		tx:         tx,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Register creates a user, a personal team and the owner membership inside a
// single transaction, then issues tokens.
func (u *authUsecase) Register(ctx context.Context, input ports.RegisterInput) (*ports.AuthResult, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if name == "" || email == "" || len(input.Password) < 8 {
		return nil, apperr.BadRequest(apperr.CodeValidation, "name, email and a password of at least 8 characters are required")
	}

	hashed, err := u.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	teamName := strings.TrimSpace(input.TeamName)
	if teamName == "" {
		teamName = name + "'s team"
	}

	var user *userdomain.User
	err = u.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		team := &teamdomain.Team{Name: teamName}
		if err := u.teams.Create(ctx, team); err != nil {
			return err
		}
		user = &userdomain.User{
			Name:          name,
			Email:         email,
			PasswordHash:  hashed,
			PrimaryTeamID: &team.ID,
		}
		if err := u.users.Create(ctx, user); err != nil {
			return err
		}
		return u.teams.AddMember(ctx, team.ID, user.ID, teamdomain.RoleOwner)
	})
	if err != nil {
		return nil, apperr.TranslateConflict(err, "email is already registered")
	}
	return u.issue(ctx, user)
}

// Login verifies credentials and issues tokens.
func (u *authUsecase) Login(ctx context.Context, input ports.LoginInput) (*ports.AuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := u.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, kernel.ErrNotFound) {
			return nil, apperr.Unauthorized("invalid email or password")
		}
		return nil, err
	}
	if err := u.hasher.Compare(user.PasswordHash, input.Password); err != nil {
		return nil, apperr.Unauthorized("invalid email or password")
	}
	return u.issue(ctx, user)
}

// Refresh rotates a valid refresh token and issues a new token pair.
func (u *authUsecase) Refresh(ctx context.Context, refreshToken string) (*ports.AuthResult, error) {
	claims, err := u.tokens.Parse(refreshToken, kernel.TokenTypeRefresh)
	if err != nil {
		return nil, apperr.Unauthorized("invalid refresh token")
	}
	valid, err := u.tokenStore.HasRefresh(ctx, claims.JTI)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, apperr.Unauthorized("refresh token has been revoked")
	}
	user, err := u.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, apperr.TranslateNotFound(err, "user")
	}
	if err := u.tokenStore.RevokeRefresh(ctx, claims.JTI); err != nil {
		return nil, err
	}
	return u.issue(ctx, user)
}

// Logout blacklists the access token and revokes the refresh token when given.
func (u *authUsecase) Logout(ctx context.Context, userID uint64, accessJTI string, accessExpiry time.Time, refreshToken string) error {
	if accessJTI != "" {
		if ttl := time.Until(accessExpiry); ttl > 0 {
			if err := u.tokenStore.BlacklistAccess(ctx, accessJTI, ttl); err != nil {
				return err
			}
		}
	}
	if refreshToken != "" {
		if claims, err := u.tokens.Parse(refreshToken, kernel.TokenTypeRefresh); err == nil {
			if err := u.tokenStore.RevokeRefresh(ctx, claims.JTI); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *authUsecase) issue(ctx context.Context, user *userdomain.User) (*ports.AuthResult, error) {
	accessClaims, access, err := u.tokens.Generate(user.ID, kernel.TokenTypeAccess, u.accessTTL)
	if err != nil {
		return nil, err
	}
	refreshClaims, refresh, err := u.tokens.Generate(user.ID, kernel.TokenTypeRefresh, u.refreshTTL)
	if err != nil {
		return nil, err
	}
	if err := u.tokenStore.SaveRefresh(ctx, refreshClaims.JTI, user.ID, u.refreshTTL); err != nil {
		return nil, err
	}
	return &ports.AuthResult{
		User:             user,
		AccessToken:      access,
		RefreshToken:     refresh,
		AccessExpiresAt:  accessClaims.Expiry,
		RefreshExpiresAt: refreshClaims.Expiry,
	}, nil
}
