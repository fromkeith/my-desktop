package middleware

import (
	"fromkeith/my-desktop-server/auth"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthTokenExtract() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			claims, err := auth.ValidateToken(strings.TrimPrefix(authHeader, "Bearer "))
			if err != nil {
				panic(err)
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
