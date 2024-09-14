package user

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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
