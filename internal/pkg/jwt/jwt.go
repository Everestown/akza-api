package jwt

import (
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	AdminID string `json:"admin_id"`
	Email   string `json:"email"`
	gojwt.RegisteredClaims
}

type Manager struct {
	secret      []byte
	expiresHours int
}

func NewManager(secret string, expiresHours int) *Manager {
	return &Manager{secret: []byte(secret), expiresHours: expiresHours}
}

// GenerateToken creates a signed JWT for the given admin.
func (m *Manager) GenerateToken(adminID, email string) (string, error) {
	claims := Claims{
		AdminID: adminID,
		Email:   email,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(time.Duration(m.expiresHours) * time.Hour)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
			Issuer:    "akza-api",
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseClaims validates a JWT and returns its claims.
func (m *Manager) ParseClaims(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
