package user

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FirstName string    `json:"firstName" validate:"required"`
	LastName  string    `json:"lastName" validate:"required"`
	Email     string    `json:"email" validate:"required,email"`
	Password  string    `json:"password" validate:"required,min=8"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

var ErrDuplicateEmail = errors.New("email already exists")

func (repo *UserRepository) AddUser(user *User) error {
	if err := repo.db.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (repo *UserRepository) GetUserByEmail(email string) (*User, error) {
	var user User
	if err := repo.db.Where("Email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepository) GetUserById(userId uint) (*User, error) {
	var user User
	err := repo.db.First(&user, userId).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepository) DeleteUser(userId uint) error {
	return repo.db.Delete(&User{}, userId).Error
}

func (repo *UserRepository) UpdateUser(user *User) error {
	return repo.db.Save(user).Error
}
