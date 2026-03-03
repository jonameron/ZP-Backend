package repository

import (
	"gorm.io/gorm"
	"zeitpass/internal/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUserID(userID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("user_id = ?", userID).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *UserRepository) FindByIDs(ids []uint) (map[uint]*models.User, error) {
	var users []models.User
	if len(ids) == 0 {
		return map[uint]*models.User{}, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	if err != nil {
		return nil, err
	}
	m := make(map[uint]*models.User, len(users))
	for i := range users {
		m[users[i].ID] = &users[i]
	}
	return m, nil
}

func (r *UserRepository) FindAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Order("first_name ASC").Find(&users).Error
	return users, err
}
