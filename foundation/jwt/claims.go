package jwt

import "github.com/dgrijalva/jwt-go"

// Claims are extended with user ID to pull User model from database after authentication
type Claims struct {
	jwt.StandardClaims
	AccountID uint64
	Type      TokenType
}
