package middlewares

import (
	"calc/foundation/jwt"
	"calc/internal/berrors"
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go/request"
	"net/http"
)

const (
	AccountIDCtxKey = iota
)

func Verify(jwtAuth *jwt.Authenticator, tokenType jwt.TokenType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			tokenStr, err := request.AuthorizationHeaderExtractor.ExtractToken(r)
			if err != nil {
				respondError(w, r, berrors.WrapWithError(ErrNoToken, err), http.StatusUnauthorized)
				return
			}

			claims, err := jwtAuth.Validate(ctx, tokenStr, tokenType)
			if err != nil {
				if errors.Is(err, jwt.ErrInvalidToken) {
					respondError(w, r, berrors.WrapWithError(ErrInvalidToken, err), http.StatusUnauthorized)
					return
				}
				respondError(w, r, err, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, AccountIDCtxKey, claims.AccountID)))
		})
	}
}
