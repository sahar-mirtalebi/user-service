package user

import (
	"errors"
	"net/http"
	"user-service/auth"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type UserHandler struct {
	service  *UserService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewUserHandler(service *UserService, logger *zap.Logger, validate *validator.Validate) *UserHandler {
	return &UserHandler{service: service, logger: logger, validate: validate}
}

func (handler *UserHandler) RegisterUser(c echo.Context) error {
	var user User
	if err := c.Bind(&user); err != nil {
		handler.logger.Error("fail to bind the request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := handler.validate.Struct(user); err != nil {
		handler.logger.Error("validation failed", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data: "+err.Error())
	}

	userId, err := handler.service.CreateUser(user)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			handler.logger.Warn("attempt to register with duplicate email", zap.String("email", user.Email))
			return echo.NewHTTPError(http.StatusConflict, "this email is already registered")
		}
		handler.logger.Error("fail to create user", zap.Error(err))
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
		handler.logger.Error("fail to bind the request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	user, err := handler.service.AthenticateUser(creds.Email, creds.Password)
	if err != nil {
		handler.logger.Error("authenticated fail", zap.Error(err))
		return err
	}

	token, err := auth.GenerateToken(user.ID, user.Email, "login")
	if err != nil {
		handler.logger.Error("failed to generate token", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "error generating token")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})
}

func (handler *UserHandler) FogotPassword(c echo.Context) error {
	var request struct {
		Email string `json:"email"`
	}

	if err := c.Bind(&request); err != nil {
		handler.logger.Error("fail to bind the request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	user, err := handler.service.CheckEmailExists(request.Email)
	if err != nil {
		handler.logger.Error("authenticated fail", zap.Error(err))
		return err
	}

	token, err := auth.GenerateToken(user.ID, user.Email, "reset")
	if err != nil {
		handler.logger.Error("failed to generate token", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "error generating token")
	}

	resetLink := "http://localhost:8080/reset-password?token=" + token

	err = handler.service.sendResetLinkEmail(user.Email, resetLink)
	if err != nil {
		handler.logger.Error("faild to send reset email", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "could not send reset email")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"resetLink": resetLink,
		"message":   "reset link sent to email",
	})
}

func (handler *UserHandler) ResetPassword(c echo.Context) error {
	token := c.QueryParam("token")

	var Request struct {
		NewPassword string `json:"newPassword"`
	}
	if err := c.Bind(&Request); err != nil {
		handler.logger.Error("fail to bind the request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		handler.logger.Error("invalid or expired token", zap.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	userIdFloat, ok := claims["UserId"].(float64)
	if !ok {
		handler.logger.Error("invalid token claim", zap.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claim")
	}
	userId := uint(userIdFloat)

	err = handler.service.UpdatePassword(userId, Request.NewPassword)
	if err != nil {
		handler.logger.Error("failed to update password")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update password")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "your password updated successfully",
	})
}

func (handler *UserHandler) DeleteAccount(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		handler.logger.Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	err := handler.service.DeleteAccount(userId)
	if err != nil {
		handler.logger.Error("failed to delete account", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete account")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "account deleted successfully",
	})

}

func (handler *UserHandler) UpdateAccount(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		handler.logger.Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var updatedUser struct {
		FirstName string `json:"firstName" validate:"omitempty,min=3"`
		LastName  string `json:"lastName" validate:"omitempty,min=3"`
		Email     string `json:"email" validate:"omitempty,email"`
	}
	if err := c.Bind(&updatedUser); err != nil {
		handler.logger.Error("fail to bind the request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := handler.validate.Struct(updatedUser); err != nil {
		handler.logger.Error("provided data is invalid", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data")
	}

	err := handler.service.UpdateAccount(userId, updatedUser)
	if err != nil {
		handler.logger.Error("failed to update account", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "error updating account")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "account updated successfully",
	})
}

func (handler *UserHandler) RetrieveAccount(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		handler.logger.Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	user, err := handler.service.RetrieveAccount(userId)
	if err != nil {
		handler.logger.Error("failed to retrieve user account", zap.Error(err))
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
