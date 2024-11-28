package user

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint
	FirstName string
	LastName  string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) AddUser(user *User) error {
	if err := repo.db.Create(&user).Error; err != nil {
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
