package httpx

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ryanpramasyana/gdc/internal/shared/apperr"
	"github.com/ryanpramasyana/gdc/internal/shared/reqctx"
)

// CurrentUserID returns the authenticated user's identifier from the request
// context, or an unauthorized error when the request is anonymous.
func CurrentUserID(c *gin.Context) (uint64, error) {
	id, ok := reqctx.UserID(c.Request.Context())
	if !ok {
		return 0, apperr.Unauthorized("authentication required")
	}
	return id, nil
}

// PathID parses an unsigned integer path parameter.
func PathID(c *gin.Context, name string) (uint64, error) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, apperr.BadRequest(apperr.CodeValidation, "invalid "+name)
	}
	return id, nil
}
