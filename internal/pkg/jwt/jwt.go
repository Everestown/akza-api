package jwt

import (
	"fmt"
	"strconv"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	AdminID string `json:"admin_id"` // stored as string in JWT for compatibility
	Email   string `json:"email"`
	gojwt.RegisteredClaims
}

type Manager struct {
	secret       []byte
	expiresHours int
}

func NewManager(secret string, expiresHours int) *Manager {
	return &Manager{secret: []byte(secret), expiresHours: expiresHours}
}

// GenerateToken creates a signed JWT for the given admin (int64 ID).
func (m *Manager) GenerateToken(adminID int64, email string) (string, error) {
	claims := Claims{
		AdminID: strconv.FormatInt(adminID, 10),
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
	if err != nil { return nil, fmt.Errorf("parse token: %w", err) }
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid { return nil, fmt.Errorf("invalid token") }
	return claims, nil
}

// ParseAdminID parses the string AdminID from claims into int64.
func ParseAdminID(adminIDStr string) (int64, error) {
	return strconv.ParseInt(adminIDStr, 10, 64)
}
