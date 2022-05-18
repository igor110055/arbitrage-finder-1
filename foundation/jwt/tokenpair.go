package jwt

type TokenType string

const (
	Access  TokenType = "ACCESS"
	Refresh TokenType = "REFRESH"
)

// TokenPair is pair of access and refresh tokens
type TokenPair struct {
	Access  string `json:"access_token" validate:"required"`
	Refresh string `json:"refresh_token" validate:"required"`
}
