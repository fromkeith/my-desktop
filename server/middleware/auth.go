package middleware

import (
	"fromkeith/my-desktop-server/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthTokenExtract() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			claims, err := auth.ValidateToken(strings.TrimPrefix(authHeader, "Bearer "))
			if err != nil {
				// TODO: don't abort, don't just error
				// handle expired credentials
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			c.Set("claims", *claims)
			c.Set("accountId", claims.Subject)
			c.Set("isAuthed", true)

		} else {
			c.Set("isAuthed", false)
		}
		c.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("isAuthed") {
			c.Abort()
			return
		}
		c.Next()
	}
}
