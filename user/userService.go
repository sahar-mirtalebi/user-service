package user

import (
	"errors"
	"strings"
	"time"
	"user-service/auth"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

var ErrDuplicateEmail = errors.New("email already exists")
var ErrUserNotFound = errors.New("user not found")

func (service *UserService) CreateUser(userDto UserDto) (*uint, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userDto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		FirstName: userDto.FirstName,
		LastName:  userDto.LastName,
		Email:     userDto.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	err = service.repo.AddUser(user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}
	return &user.ID, nil
}

func (service *UserService) AthenticateUser(email, password string) (*string, error) {
	user, err := service.repo.GetUserByEmail(email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	token, err := auth.GenerateToken(user.ID, user.Email, "login")
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (service *UserService) FogotPassword(email string) (*string, error) {
	user, err := service.CheckEmailExists(email)
	if err != nil {
		return nil, err
	}

	token, err := auth.GenerateToken(user.ID, user.Email, "reset")
	if err != nil {
		return nil, err
	}

	resetLink := "http://localhost:8080/reset-password?token=" + token

	err = service.sendResetLinkEmail(email, resetLink)
	if err != nil {
		return nil, err
	}

	return &resetLink, nil
}

func (service *UserService) CheckEmailExists(email string) (*User, error) {
	user, err := service.repo.GetUserByEmail(email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return user, nil
}

func (service *UserService) sendResetLinkEmail(email, resetLink string) error {
	zap.L().Info("sending reset link email", zap.String("email", email), zap.String("resetLink", resetLink))
	return nil
}

func (service *UserService) UpdatePassword(token, newPassword string) error {
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		return err
	}

	userIdFloat, ok := claims["UserId"].(float64)
	if !ok {
		return err
	}

	userId := uint(userIdFloat)

	user, err := service.repo.GetUserById(userId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return err
	}

	user.Password = string(hashedNewPassword)
	user.UpdatedAt = time.Now()

	err = service.repo.UpdateUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (service *UserService) DeleteAccount(userId uint) error {
	err := service.repo.DeleteUser(userId)
	if err != nil {
		return err
	}
	return nil
}

func (service *UserService) UpdateAccount(userId uint, updatedUser struct {
	FirstName string `json:"firstName" validate:"omitempty,min=3"`
	LastName  string `json:"lastName" validate:"omitempty,min=3"`
	Email     string `json:"email" validate:"omitempty,email"`
}) error {
	user, err := service.repo.GetUserById(userId)
	if err != nil {
		return err
	}

	if updatedUser.FirstName != "" {
		user.FirstName = updatedUser.FirstName
	}
	if updatedUser.LastName != "" {
		user.LastName = updatedUser.LastName
	}
	if updatedUser.Email != "" {
		user.Email = updatedUser.Email
	}
	user.UpdatedAt = time.Now()

	if err := service.repo.UpdateUser(user); err != nil {
		return err
	}

	return nil
}

func (service *UserService) RetrieveAccount(userId uint) (*User, error) {
	user, err := service.repo.GetUserById(userId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}
