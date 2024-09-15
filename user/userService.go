package user

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo   *UserRepository
	logger *zap.Logger
}

func NewUserService(repo *UserRepository, logger *zap.Logger) *UserService {
	return &UserService{repo: repo, logger: logger}
}

func (service *UserService) CreateUser(user User) (uint, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		service.logger.Error("failed to hash password", zap.Error(err))
		return 0, echo.NewHTTPError(http.StatusInternalServerError, "error hashing password")
	}

	user.Password = string(hashedPassword)
	user.CreatedAt = time.Now()

	err = service.repo.AddUser(&user)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			return 0, ErrDuplicateEmail
		}
		service.logger.Error("fail to create user", zap.Error(err))
		return 0, err
	}
	return user.ID, nil
}

func (service *UserService) AthenticateUser(email, password string) (*User, error) {
	user, err := service.repo.GetUserByEmail(email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
		}
		service.logger.Error("error retrieving user", zap.Error(err))
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
	}

	return user, nil
}

func (service *UserService) CheckEmailExists(email string) (*User, error) {
	user, err := service.repo.GetUserByEmail(email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
		}
		service.logger.Error("failed to find user", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "error retrieving user")
	}
	return user, nil
}

func (service *UserService) sendResetLinkEmail(email, resetLink string) error {
	service.logger.Info("sending reset link email", zap.String("email", email), zap.String("resetLink", resetLink))
	return nil
}

func (service *UserService) UpdatePassword(userId uint, newPassword string) error {
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		service.logger.Error("failed to hash password", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "error hashing password")
	}

	user, err := service.repo.GetUserById(userId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			service.logger.Warn("user not found", zap.Uint("userId", userId))
			echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		service.logger.Error("failed to find user", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "error retrieving user")
	}

	user.Password = string(hashedNewPassword)

	err = service.repo.UpdateUser(user)
	if err != nil {
		service.logger.Error("failed to update user password", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "error updating password")
	}

	return nil
}
