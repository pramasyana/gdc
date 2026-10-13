// Package security provides password hashing and JWT token handling.
package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ryanpramasyana/gdc/internal/shared/kernel"
)

// jwtClaims is the signed token payload.
type jwtClaims struct {
	UserID    uint64 `json:"uid"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

// JWTProvider signs and parses JWT tokens using HMAC-SHA256.
type JWTProvider struct {
	secret []byte
	issuer string
}

// NewJWTProvider builds a JWTProvider.
func NewJWTProvider(secret, issuer string) *JWTProvider {
	return &JWTProvider{secret: []byte(secret), issuer: issuer}
}

// Generate implements kernel.TokenProvider.
func (p *JWTProvider) Generate(userID uint64, tokenType kernel.TokenType, ttl time.Duration) (*kernel.TokenClaims, string, error) {
	now := time.Now()
	jti := uuid.NewString()
	claims := jwtClaims{
		UserID:    userID,
		TokenType: string(tokenType),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Issuer:    p.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(p.secret)
	if err != nil {
		return nil, "", fmt.Errorf("security: sign token: %w", err)
	}
	return &kernel.TokenClaims{UserID: userID, JTI: jti, Type: tokenType, Expiry: claims.ExpiresAt.Time}, signed, nil
}

// Parse implements kernel.TokenProvider.
func (p *JWTProvider) Parse(token string, tokenType kernel.TokenType) (*kernel.TokenClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("security: unexpected signing method")
		}
		return p.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("security: parse token: %w", err)
	}
	claims, ok := parsed.Claims.(*jwtClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("security: invalid token")
	}
	if claims.TokenType != string(tokenType) {
		return nil, fmt.Errorf("security: unexpected token type %q", claims.TokenType)
	}
	return &kernel.TokenClaims{
		UserID: claims.UserID,
		JTI:    claims.ID,
		Type:   kernel.TokenType(claims.TokenType),
		Expiry: claims.ExpiresAt.Time,
	}, nil
}

// BcryptHasher hashes passwords with bcrypt.
type BcryptHasher struct {
	cost int
}

// NewPasswordHasher builds a BcryptHasher using the default cost.
func NewPasswordHasher() *BcryptHasher {
	return &BcryptHasher{cost: bcrypt.DefaultCost}
}

// Hash implements kernel.PasswordHasher.
func (h *BcryptHasher) Hash(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", fmt.Errorf("security: hash password: %w", err)
	}
	return string(hashed), nil
}

// Compare implements kernel.PasswordHasher.
func (h *BcryptHasher) Compare(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
