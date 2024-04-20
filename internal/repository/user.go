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
	GetUserWithRolesAndPermission(ctx context.Context, userId string, permissionType string, sort string) (*[]LoginedUser, error)
	GetUserDefaultSeed(ctx context.Context, user *model.User) error
}

func NewUserRepository(r *Repository) UserRepository {
	return &userRepository{
		Repository: r,
	}
}

// 当前登录用户菜单表
type LoginedUser struct {
	Email          string `json:"email"`
	UserId         string `json:"user_id"`
	RoleId         uint   `json:"role_id"`
	RoleLabel      string `json:"role_label"`
	RoleName       string `json:"role_name"`
	PermissionId   uint   `json:"permission_id"`
	PermissionType string `json:"permission_type"`
	PermissionName string `json:"permission_name"`
	Route          string `json:"route"`
	RouteFile      string `json:"route_file"`
	Level          uint   `json:"level"`
	Sort           string `json:"sort"`
	ParentId       uint   `json:"parent_id"`
	Path           string `json:"path"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
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

func (r *userRepository) GetUserWithRolesAndPermission(ctx context.Context, userId string, permissionType string, sort string) (*[]LoginedUser, error) {
	if sort != "asc" && sort != "desc" {
		sort = "asc"
	}

	// TODO 左连接查出当前user的所拥有的权限
	var users []LoginedUser
	connect := r.DB(ctx).Table("users").Select("users.user_id, users.email, user_role.role_id, role.id, "+
		"role.role_name, role.role_label, role_permissions.permission_id, permission.permission_type, permission.route, "+
		"permission.route_file, permission.level, permission.sort, permission.parent_id, permission.path, permission.created_at, "+
		"permission.updated_at, permission.permission_name").
		Joins("left join user_role on users.user_id = user_role.user_user_id").
		Joins("left join role on user_role.role_id = role.id").
		Joins("left join role_permissions on role.id = role_permissions.role_id").
		Joins("left join permission on role_permissions.permission_id = permission.id").
		Where("users.user_id = ? AND permission.permission_type = ?", userId, permissionType).
		Scan(&users)
	// 执行查询语句时的出现异常
	if connect.Error != nil {
		r.logger.WithContext(ctx).Error("databaseError!", zap.Error(connect.Error))
		return nil, connect.Error
	}
	// End

	// 记录为空
	if len(users) == 0 {
		return nil, v1.ErrEmptyRecord
	}

	return &users, nil

	//if err := r.DB(ctx).Preload("Roles").Preload("Roles.Permissions", func(db *gorm.DB) *gorm.DB {
	//	return db.Order("sort " + sort)
	//}).Where("user_id = ?", userId).First(&user).Error; err != nil {
	//	if errors.Is(err, gorm.ErrRecordNotFound) {
	//		return nil, v1.ErrEmptyRecord
	//	}
	//	return nil, err
	//}
	//return &user, nil
}
