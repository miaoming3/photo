package service

import (
	"errors"

	"github.com/miaoming3/photo/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	DB *gorm.DB
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{DB: db}
}

// CreateUser 创建新用户
func (s *UserService) CreateUser(user *model.User) error {
	// 检查用户名是否已存在
	var existingUser model.User
	if err := s.DB.Where("username = ?", user.Username).First(&existingUser).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if err := s.DB.Where("email = ?", user.Email).First(&existingUser).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("邮箱已被注册")
	}

	// 密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)

	// 创建用户
	return s.DB.Create(user).Error
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := s.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// VerifyPassword 验证密码
func (s *UserService) VerifyPassword(user *model.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}
