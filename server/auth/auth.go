package auth

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type DesktopClaims struct {
	jwt.RegisteredClaims
}

// be sure to have set the Subject
func CreateToken(claims DesktopClaims) (string, error) {
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

func ValidateToken(tokenStr string) (*DesktopClaims, error) {
	var token, err = jwt.ParseWithClaims(tokenStr, &DesktopClaims{}, keyFunc)
	if err != nil {
		return nil, err
	}

	// Type assert the claims
	if claims, ok := token.Claims.(*DesktopClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("Invalid claims")
}

func ClaimsOrNil(ctx context.Context) *DesktopClaims {
	claimsI := ctx.Value("claims")
	if claimsI != nil {
		claims := claimsI.(DesktopClaims)
		return &claims
	}
	return nil
}
