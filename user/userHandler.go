package user

import (
	"errors"
	"net/http"

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
