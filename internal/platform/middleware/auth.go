package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/ryanpramasyana/gdc/internal/shared/apperr"
	"github.com/ryanpramasyana/gdc/internal/shared/kernel"
	"github.com/ryanpramasyana/gdc/internal/shared/reqctx"
	"github.com/ryanpramasyana/gdc/internal/shared/response"
)

// Auth validates the bearer access token, rejects revoked tokens and stores
// the caller identity on the request context.
func Auth(tokens kernel.TokenProvider, tokenStore kernel.TokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearer(c)
		if token == "" {
			response.Error(c, apperr.Unauthorized("missing bearer token"))
			return
		}

		claims, err := tokens.Parse(token, kernel.TokenTypeAccess)
		if err != nil {
			response.Error(c, apperr.Unauthorized("invalid or expired token"))
			return
		}

		blacklisted, err := tokenStore.IsAccessBlacklisted(c.Request.Context(), claims.JTI)
		if err != nil {
			response.Error(c, err)
			return
		}
		if blacklisted {
			response.Error(c, apperr.Unauthorized("token has been revoked"))
			return
		}

		c.Request = c.Request.WithContext(reqctx.WithUserID(c.Request.Context(), claims.UserID))
		c.Set(ContextAccessJTI, claims.JTI)
		c.Set(ContextAccessExp, claims.Expiry)
		c.Next()
	}
}
