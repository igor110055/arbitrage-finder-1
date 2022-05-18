package jwt

import (
	"calc/foundation/id"
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/oklog/ulid/v2"
	"github.com/pkg/errors"
	"io/ioutil"
	"time"
)

// Authenticator generates, validates and revokes users' access and refresh tokens.
type Authenticator struct {
	algorithm                       jwt.SigningMethod
	privateKey                      interface{}
	publicKey                       interface{}
	parser                          *jwt.Parser
	accessLifetime, refreshLifetime time.Duration
	rtKeeper                        RefreshTokenKeeper
}

// New creates Authenticator
func New(
	algorithm jwt.SigningMethod,
	privateKey, publicKey interface{},
	options ...AuthenticatorOption,
) *Authenticator {
	auth := &Authenticator{
		algorithm:  algorithm,
		privateKey: privateKey,
		publicKey:  publicKey,
		parser:     &jwt.Parser{ValidMethods: []string{algorithm.Alg()}},
	}

	for _, option := range options {
		option(auth)
	}

	return auth
}

// NewFromFiles creates Authenticator from public and private keys filepaths
func NewFromFiles(
	alg, privateKeyFilepath, publicKeyFilepath string,
	options ...AuthenticatorOption,
) (*Authenticator, error) {
	algorithm := jwt.GetSigningMethod(alg)
	if algorithm == nil {
		return nil, errors.Errorf("unknown signing method %q", alg)
	}

	privatePEM, err := ioutil.ReadFile(privateKeyFilepath)
	if err != nil {
		return nil, errors.Wrap(err, "reading auth private key")
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return nil, errors.Wrap(err, "parsing auth private key")
	}

	publicPEM, err := ioutil.ReadFile(publicKeyFilepath)
	if err != nil {
		return nil, errors.Wrap(err, "reading auth public sessionID")
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicPEM)
	if err != nil {
		return nil, errors.Wrap(err, "parsing auth public sessionID")
	}

	return New(algorithm, privateKey, publicKey, options...), nil
}

// GenerateTokenPair creates token pair for user who requires it
func (a *Authenticator) GenerateTokenPair(ctx context.Context, accountID uint64) (*TokenPair, error) {
	// Generate access and refresh tokens as future pair
	_, accessToken, err := a.generateToken(accountID, Access)
	if err != nil {
		return nil, err
	}
	refreshTokenID, refreshToken, err := a.generateToken(accountID, Refresh)
	if err != nil {
		return nil, err
	}

	if a.rtKeeper != nil {
		if err := a.rtKeeper.Save(ctx, accountID, refreshTokenID); err != nil {
			return nil, err
		}
	}

	return &TokenPair{Access: accessToken, Refresh: refreshToken}, nil
}

// Validate token with TokenType
func (a *Authenticator) Validate(
	ctx context.Context,
	token string,
	tokenType TokenType,
) (*Claims, error) {
	var claims Claims
	// Basic token validation (expiration, signing method etc)
	if _, err := a.parser.ParseWithClaims(token, &claims, func(*jwt.Token) (interface{}, error) {
		return a.publicKey, nil
	}); err != nil {
		return nil, errors.Wrap(ErrInvalidToken, err.Error())
	}
	if claims.Type != tokenType {
		err := errors.Errorf("invalid token type %q, %q is expected", claims.Type, tokenType)
		return nil, errors.Wrap(ErrInvalidToken, err.Error())
	}

	// Check whether refresh token exists and delete it then
	if a.rtKeeper != nil && claims.Type == Refresh {
		tokenID, err := ulid.Parse(claims.Id)
		if err != nil {
			return nil, errors.Wrap(ErrInvalidToken, err.Error())
		}

		if err := a.rtKeeper.Delete(ctx, tokenID); err != nil {
			return nil, err
		}
	}

	return &claims, nil
}

// generateToken for user with specific token lifetime that depends on TokenType
// Returns token ID and token string representation
func (a *Authenticator) generateToken(accountID uint64, tokenType TokenType) (ulid.ULID, string, error) {
	lifetime := a.accessLifetime
	if tokenType == Refresh {
		lifetime = a.refreshLifetime
	}

	tokenID := id.ULID()

	token := jwt.NewWithClaims(a.algorithm, Claims{
		StandardClaims: jwt.StandardClaims{
			Id:        tokenID.String(),
			ExpiresAt: time.Now().Add(lifetime).Unix(),
		},
		AccountID: accountID,
		Type:      tokenType,
	})
	tokenStr, err := token.SignedString(a.privateKey)
	return tokenID, tokenStr, errors.Wrap(err, "failed to sign token")
}
