package usecase_test

import (
	"testing"
	"time"

	"github.com/ryanpramasyana/gdc/internal/package/auth/ports"
	"github.com/ryanpramasyana/gdc/internal/package/auth/usecase"
	"github.com/ryanpramasyana/gdc/test/mocks"
)

const (
	testAccessTTL  = 15 * time.Minute
	testRefreshTTL = 168 * time.Hour
)

type authFixture struct {
	uc         ports.AuthUsecase
	store      *mocks.Store
	tokenStore *mocks.TokenStore
}

func newAuthFixture(t *testing.T) *authFixture {
	t.Helper()
	store := mocks.NewStore()
	tokenStore := mocks.NewTokenStore()
	uc := usecase.New(
		mocks.NewUserRepository(store),
		mocks.NewTeamRepository(store),
		mocks.NewPasswordHasher(),
		mocks.NewTokenProvider(),
		tokenStore,
		mocks.NewTxManager(store),
		testAccessTTL,
		testRefreshTTL,
	)
	return &authFixture{uc: uc, store: store, tokenStore: tokenStore}
}
