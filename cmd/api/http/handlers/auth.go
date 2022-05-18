package handlers

import (
	"calc/cmd/api/http/handlers/requests"
	"calc/cmd/api/http/handlers/responses"
	"calc/cmd/api/http/middlewares"
	"calc/internal/berrors"
	"calc/internal/services/auth"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type authGroup struct {
	authService *auth.Service
}

func newAuthGroup(authService *auth.Service) *authGroup {
	return &authGroup{
		authService: authService,
	}
}

// SignUp godoc
// @Tags Auth
// @Router /auth/sign-up [post]
// @Summary Sign up with phone and password
// @Description Afterwards user receives code on his phone  provided
// @Accept json
// @Produce json
// @Param body body requests.SignUp true " "
// @Success 200 {object} responses.SignUp
// @Failure 400 {object} berrors.BusinessError
// @Failure 500
func (ag *authGroup) SignUp(r *http.Request) (interface{}, error) {
	var req requests.SignUp
	if err := requests.Bind(r, &req); err != nil {
		fmt.Println(berrors.WrapWithError(auth.ErrInvalidInput, err))
		return nil, berrors.WrapWithError(auth.ErrInvalidInput, err)
	}

	id, err := ag.authService.SignUp(r.Context(), &auth.SignUpArgs{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return responses.SignUp{
		PhoneConfirmationID: id,
	}, nil
}

// Confirm godoc
// @Tags Auth
// @Router /auth/confirm [post]
// @Summary Confirm sign up
// @Description User will be created as the result of this action, registration is complete
// @Accept json
// @Produce json
// @Param body body requests.Confirm true " "
// @Success 200 {object} jwt.TokenPair
// @Failure 400 {object} berrors.BusinessError
// @Failure 500
func (ag *authGroup) Confirm(r *http.Request) (interface{}, error) {
	var req requests.Confirm
	if err := requests.Bind(r, &req); err != nil {
		return nil, berrors.WrapWithError(auth.ErrInvalidInput, err)
	}

	return ag.authService.Confirm(r.Context(), &auth.ConfirmArgs{
		ConfirmationID: req.ConfirmationID,
		Password:       req.Password,
		Code:           req.Code,
	})
}

// Code godoc
// @Tags Auth
// @Router /auth/code/{phone} [post]
// @Summary Send code
// @Description Send releans with confirmation sign up code
// @Produce json
// @Param phone path string true "Phone"
// @Success 200 {object} responses.SignUp
// @Failure 400 {object} berrors.BusinessError
// @Failure 500
func (ag *authGroup) Code(r *http.Request) (interface{}, error) {
	var req requests.Code
	if err := requests.Bind(r, &req); err != nil {
		return nil, berrors.WrapWithError(auth.ErrInvalidInput, err)
	}

	id, err := ag.authService.CreatePhoneConfirmation(r.Context(), req.Phone)
	if err != nil {
		return nil, err
	}

	return responses.SignUp{
		PhoneConfirmationID: id,
	}, nil
}

// SignIn godoc
// @Tags Auth
// @Router /auth/sign-in [post]
// @Summary Sign in with phone and password
// @Description Responds with generated token pair as successful result
// @Accept json
// @Produce json
// @Param body body requests.SignIn true " "
// @Success 200 {object} jwt.TokenPair
// @Failure 400 {object} berrors.BusinessError
// @Failure 500
func (ag *authGroup) SignIn(r *http.Request) (interface{}, error) {
	var req requests.SignIn
	if err := requests.Bind(r, &req); err != nil {
		return nil, berrors.WrapWithError(auth.ErrInvalidInput, err)
	}

	return ag.authService.SignIn(r.Context(), req.Phone, req.Password)
}

// Refresh godoc
// @Tags Auth
// @Router /auth/refresh [post]
// @Security JWT-Token
// @Summary Refresh both access and refresh tokens
// @Produce json
// @Success 200 {object} jwt.TokenPair
// @Success 400 {object} berrors.BusinessError
// @Failure 500
func (ag *authGroup) Refresh(r *http.Request) (interface{}, error) {
	accountID := r.Context().Value(middlewares.AccountIDCtxKey)
	if accountID == nil {
		return nil, errors.New("no current account")
	}

	pair, err := ag.authService.Refresh(r.Context(), accountID.(uint64))
	if err != nil {
		return nil, err
	}

	return pair, nil
}

// CheckPhone godoc
// @Tags Auth
// @Router /auth/check/{phone} [get]
// @Summary Check phone: validate and check user by this phone
// @Produce json
// @Param phone path string true "Phone"
// @Success 200 {object} responses.UserExists
// @Failure 400 {object} berrors.BusinessError
// @Failure 500
func (ag *authGroup) CheckPhone(r *http.Request) (interface{}, error) {
	var req requests.CheckPhone
	if err := requests.Bind(r, &req); err != nil {
		return nil, berrors.WrapWithError(auth.ErrInvalidInput, err)
	}

	exists, err := ag.authService.CheckPhone(r.Context(), req.Phone)
	if err != nil {
		return nil, err
	}

	return responses.UserExists{
		Exists: exists,
	}, nil
}
