package service

import (
	"errors"
	"user-service/internal/model"
	"user-service/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(user *model.User) (*model.User, error) {
	// 检查邮箱是否已存在
	if existingUser, _ := s.repo.GetByEmail(user.Email); existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// TODO: 密码加密

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser 获取用户
func (s *UserService) GetUser(id uint) (*model.User, error) {
	return s.repo.GetByID(id)
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(id uint, user *model.User) (*model.User, error) {
	existingUser, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	existingUser.Username = user.Username
	existingUser.Email = user.Email
	if user.Password != "" {
		// TODO: 密码加密
		existingUser.Password = user.Password
	}

	if err := s.repo.Update(existingUser); err != nil {
		return nil, err
	}

	return existingUser, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}
