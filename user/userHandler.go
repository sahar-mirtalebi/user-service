package user

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type UserDto struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

type UserHandler struct {
	service  *UserService
	validate *validator.Validate
}

func NewUserHandler(service *UserService, validate *validator.Validate) *UserHandler {
	return &UserHandler{service: service, validate: validate}
}

func (handler *UserHandler) RegisterUser(c echo.Context) error {
	var user UserDto
	if err := c.Bind(&user); err != nil {
		zap.L().Error("fail to bind the request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := handler.validate.Struct(user); err != nil {
		zap.L().Error("validation failed", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data: "+err.Error())
	}

	userId, err := handler.service.CreateUser(user)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			zap.L().Warn("attempt to register with duplicate email", zap.String("email", user.Email))
			return echo.NewHTTPError(http.StatusConflict, "this email is already registered")
		}
		zap.L().Error("fail to create user", zap.Error(err))
		return err
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"user_id": userId,
	})
}

func (handler *UserHandler) LoginUser(c echo.Context) error {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.Bind(&creds); err != nil {
		zap.L().Error("fail to bind the request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	token, err := handler.service.AthenticateUser(creds.Email, creds.Password)
	if err != nil {
		zap.L().Error("authenticated fail", zap.Error(err))
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})
}

func (handler *UserHandler) ForgotPassword(c echo.Context) error {
	var request struct {
		Email string `json:"email"`
	}

	if err := c.Bind(&request); err != nil {
		zap.L().Error("fail to bind the request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	resetLink, err := handler.service.FogotPassword(request.Email)
	if err != nil {
		zap.L().Error("fail to generate reset email")
		return echo.NewHTTPError(http.StatusInternalServerError, "error generating reset email")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"resetLink": resetLink,
	})
}

func (handler *UserHandler) ResetPassword(c echo.Context) error {
	token := c.QueryParam("token")

	var Request struct {
		NewPassword string `json:"newPassword"`
	}
	if err := c.Bind(&Request); err != nil {
		zap.L().Error("fail to bind the request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	err := handler.service.UpdatePassword(token, Request.NewPassword)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			zap.L().Error("user does not exist", zap.Error(err))
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		zap.L().Error("failed to update password")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "your password updated successfully",
	})
}

func (handler *UserHandler) DeleteAccount(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		zap.L().Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	err := handler.service.DeleteAccount(userId)
	if err != nil {
		zap.L().Error("failed to delete account", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete account")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "account deleted successfully",
	})

}

func (handler *UserHandler) UpdateAccount(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		zap.L().Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var updatedUser struct {
		FirstName string `json:"firstName" validate:"omitempty,min=3"`
		LastName  string `json:"lastName" validate:"omitempty,min=3"`
		Email     string `json:"email" validate:"omitempty,email"`
	}

	if err := c.Bind(&updatedUser); err != nil {
		zap.L().Error("fail to bind the request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := handler.validate.Struct(updatedUser); err != nil {
		zap.L().Error("provided data is invalid", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data")
	}

	err := handler.service.UpdateAccount(userId, updatedUser)
	if err != nil {
		zap.L().Error("failed to update account", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "error updating account")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "account updated successfully",
	})
}

func (handler *UserHandler) RetrieveAccount(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		zap.L().Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	user, err := handler.service.RetrieveAccount(userId)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			zap.L().Error("user does not exist", zap.Error(err))
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		zap.L().Error("failed to retrieve user account", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "error retrieving account")
	}

	var retrievedUser struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Email     string `json:"email"`
	}

	retrievedUser.FirstName = user.FirstName
	retrievedUser.LastName = user.LastName
	retrievedUser.Email = user.Email

	return c.JSON(http.StatusOK, retrievedUser)
}
