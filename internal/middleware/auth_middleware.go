package middleware

import (
	"crypto/subtle"
	"go-inventory-reservations/internal/apperror"
	apimodel "go-inventory-reservations/internal/model/api"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// BearerTokenAuth returns middleware that validates Authorization Bearer token.
func BearerTokenAuth(expectedToken string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			unauthorizedResponse(ctx, "Authorization header is required")
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			unauthorizedResponse(ctx, "Authorization header must use Bearer token")
			return
		}

		providedToken := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
		if providedToken == "" {
			unauthorizedResponse(ctx, "Bearer token is required")
			return
		}

		if subtle.ConstantTimeCompare([]byte(providedToken), []byte(expectedToken)) != 1 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, apimodel.APIResponse{
				Success: false,
				Message: "Invalid token",
				Error: &apimodel.ErrorObject{
					Code:      apperror.CodeUnauthorizedCode,
					RequestID: requestIDFromContext(ctx),
				},
			})
			return
		}

		ctx.Next()
	}
}
