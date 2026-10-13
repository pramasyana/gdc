package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ryanpramasyana/gdc/internal/package/auth/ports"
	"github.com/ryanpramasyana/gdc/internal/shared/apperr"
)

func TestAuthRegisterCreatesUserAndTeam(t *testing.T) {
	fixture := newAuthFixture(t)

	result, err := fixture.uc.Register(context.Background(), ports.RegisterInput{
		Name:     "Jane Doe",
		Email:    "Jane@Example.com",
		Password: "password123",
	})
	require.NoError(t, err)
	require.NotNil(t, result.User.PrimaryTeamID)
	require.NotEmpty(t, result.AccessToken)
	require.NotEmpty(t, result.RefreshToken)
	require.Equal(t, "jane@example.com", result.User.Email)
}

func TestAuthRegisterDuplicateEmail(t *testing.T) {
	fixture := newAuthFixture(t)
	ctx := context.Background()
	input := ports.RegisterInput{Name: "Jane", Email: "jane@example.com", Password: "password123"}

	_, err := fixture.uc.Register(ctx, input)
	require.NoError(t, err)

	_, err = fixture.uc.Register(ctx, input)
	require.Equal(t, apperr.CodeConflict, apperr.From(err).Code)
}

// TestAuthLogin covers credential outcomes with a map of cases.
func TestAuthLogin(t *testing.T) {
	fixture := newAuthFixture(t)
	ctx := context.Background()
	_, err := fixture.uc.Register(ctx, ports.RegisterInput{Name: "Jane", Email: "jane@example.com", Password: "password123"})
	require.NoError(t, err)

	cases := map[string]struct {
		email       string
		password    string
		wantErrCode string
	}{
		"valid credentials": {email: "jane@example.com", password: "password123"},
		"wrong password":    {email: "jane@example.com", password: "wrong", wantErrCode: apperr.CodeUnauthorized},
		"unknown email":     {email: "ghost@example.com", password: "password123", wantErrCode: apperr.CodeUnauthorized},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, err := fixture.uc.Login(ctx, ports.LoginInput{Email: tc.email, Password: tc.password})
			if tc.wantErrCode != "" {
				require.Equal(t, tc.wantErrCode, apperr.From(err).Code)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, result.AccessToken)
		})
	}
}

func TestAuthRefreshRotatesToken(t *testing.T) {
	fixture := newAuthFixture(t)
	ctx := context.Background()
	registered, err := fixture.uc.Register(ctx, ports.RegisterInput{Name: "Jane", Email: "jane@example.com", Password: "password123"})
	require.NoError(t, err)

	refreshed, err := fixture.uc.Refresh(ctx, registered.RefreshToken)
	require.NoError(t, err)
	require.NotEqual(t, registered.RefreshToken, refreshed.RefreshToken)

	_, err = fixture.uc.Refresh(ctx, registered.RefreshToken)
	require.Equal(t, apperr.CodeUnauthorized, apperr.From(err).Code)
}

func TestAuthLogoutBlacklistsAccessToken(t *testing.T) {
	fixture := newAuthFixture(t)
	ctx := context.Background()
	registered, err := fixture.uc.Register(ctx, ports.RegisterInput{Name: "Jane", Email: "jane@example.com", Password: "password123"})
	require.NoError(t, err)

	err = fixture.uc.Logout(ctx, registered.User.ID, "jti-1", time.Now().Add(time.Hour), registered.RefreshToken)
	require.NoError(t, err)

	blacklisted, err := fixture.tokenStore.IsAccessBlacklisted(ctx, "jti-1")
	require.NoError(t, err)
	require.True(t, blacklisted)
}
