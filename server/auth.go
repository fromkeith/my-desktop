package main

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type desktopClaims struct {
	jwt.RegisteredClaims
}

// be sure to have set the Subject
func CreateToken(claims desktopClaims) (string, error) {
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour * 42))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())
	claims.NotBefore = jwt.NewNumericDate(time.Now())
	claims.Issuer = "localhost"
	// TODO: set ID
	// Subject should be set to userid
	// TODO: use better auth
	var token = jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	key := os.Getenv("JWT_KEY")
	return token.SignedString([]byte(key))
}

func keyFunc(tok *jwt.Token) (any, error) {
	if tok.Method != jwt.SigningMethodHS512 {
		return nil, jwt.ErrInvalidKey
	}
	return []byte(os.Getenv("JWT_KEY")), nil
}

func validateToken(tokenStr string) (*desktopClaims, error) {
	var token, err = jwt.ParseWithClaims(tokenStr, &desktopClaims{}, keyFunc)
	if err != nil {
		return nil, err
	}

	// Type assert the claims
	if claims, ok := token.Claims.(*desktopClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("Invalid claims")
}

func AuthTokenExtract() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth != "" {
			claims, err := validateToken(strings.TrimPrefix(auth, "Bearer "))
			if err != nil {
				panic(err)
			}
			c.Set("claims", claims)
		}
		c.Next()
	}
}
