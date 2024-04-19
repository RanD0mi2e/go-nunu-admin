package repository

import (
	v1 "admin-webrtc-go/api/v1"
	"admin-webrtc-go/internal/model"
	"context"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserWithRolesAndPermission(ctx context.Context, userId string, sort string) (*model.User, error)
	GetUserDefaultSeed(ctx context.Context, user *model.User) error
}

func NewUserRepository(r *Repository) UserRepository {
	return &userRepository{
		Repository: r,
	}
}

type userRepository struct {
	*Repository
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.DB(ctx).Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	if err := r.DB(ctx).Save(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, userId string) (*model.User, error) {
	var user model.User
	if err := r.DB(ctx).Where("user_id = ?", userId).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.DB(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserDefaultSeed 初始化数据库
func (r *userRepository) GetUserDefaultSeed(ctx context.Context, user *model.User) error {
	// 初始化permission
	permission := model.Permission{
		PermissionName: "默认权限",
		PermissionType: "api",
		Path:           "auth_0",
		Method:         "all",
	}
	if err := r.DB(ctx).Debug().FirstOrCreate(&permission, model.Permission{Path: permission.Path}).Error; err != nil {
		r.logger.WithContext(ctx).Error("init default Permission failed!", zap.Error(err))
		return err
	}

	// 初始化role
	role := model.Role{
		RoleLabel: "normal",
		RoleName:  "普通用户",
		Permissions: []model.Permission{
			permission,
		},
	}
	if err := r.DB(ctx).FirstOrCreate(&role, model.Role{RoleLabel: role.RoleLabel}).Error; err != nil {
		r.logger.WithContext(ctx).Error("init default Role failed!", zap.Error(err))
		return err
	}

	// 初始化User
	user.Roles = append(user.Roles, role)
	if err := r.DB(ctx).Create(&user).Error; err != nil {
		r.logger.WithContext(ctx).Error("init default User failed!", zap.Error(err))
		return err
	}

	return nil
}

func (r *userRepository) GetUserWithRolesAndPermission(ctx context.Context, userId string, sort string) (*model.User, error) {
	var user model.User
	if sort != "asc" && sort != "desc" {
		sort = "asc"
	}
	if err := r.DB(ctx).Preload("Roles").Preload("Roles.Permissions", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort " + sort)
	}).Where("user_id = ?", userId).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrEmptyRecord
		}
		return nil, err
	}
	return &user, nil
}
