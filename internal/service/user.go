package service

import (
	v1 "admin-webrtc-go/api/v1"
	"admin-webrtc-go/internal/model"
	"admin-webrtc-go/internal/repository"
	"context"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req *v1.RegisterRequest) error
	Login(ctx context.Context, req *v1.LoginRequest) (string, error)
	GetProfile(ctx context.Context, userId string) (*v1.GetProfileResponseData, error)
	UpdateProfile(ctx context.Context, userId string, req *v1.UpdateProfileRequest) error
	CheckAPIAuthPermission(ctx context.Context, userId string, api string) (bool, error)
	GetMenuTreeByUserAuth(ctx context.Context, userId string) (*v1.GetMenuTreeResponseData, error)
}

func NewUserService(service *Service, userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
		Service:  service,
	}
}

type userService struct {
	userRepo repository.UserRepository
	*Service
}

func (s *userService) Register(ctx context.Context, req *v1.RegisterRequest) error {
	// check username
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return v1.ErrInternalServerError
	}
	if user != nil {
		return v1.ErrEmailAlreadyUse
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// Generate user ID
	userId, err := s.sid.GenString()
	if err != nil {
		return err
	}
	user = &model.User{
		UserId:   userId,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	err = s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.userRepo.GetUserDefaultSeed(ctx, user); err != nil {
			return err
		}
		return nil
	})

	// Transaction demo
	//err = s.tm.Transaction(ctx, func(ctx context.Context) error {
	//	// Create a user
	//	if err = s.userRepo.Create(ctx, user); err != nil {
	//		return err
	//	}
	//	// TODO: other repo
	//	return nil
	//})
	return err
}

func (s *userService) Login(ctx context.Context, req *v1.LoginRequest) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return "", v1.ErrUnauthorized
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", err
	}
	token, err := s.jwt.GenToken(user.UserId, time.Now().Add(time.Hour*24*90))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *userService) GetProfile(ctx context.Context, userId string) (*v1.GetProfileResponseData, error) {
	user, err := s.userRepo.GetByID(ctx, userId)
	if err != nil {
		return nil, err
	}

	return &v1.GetProfileResponseData{
		UserId:   user.UserId,
		Nickname: user.Nickname,
	}, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userId string, req *v1.UpdateProfileRequest) error {
	user, err := s.userRepo.GetByID(ctx, userId)
	if err != nil {
		return err
	}

	user.Email = req.Email
	user.Nickname = req.Nickname

	if err = s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *userService) CheckAPIAuthPermission(ctx context.Context, userId string, api string) (bool, error) {
	user, err := s.userRepo.GetUserWithRolesAndPermission(ctx, userId)
	// 数据库查不到用户的权限，返回false
	if err != nil {
		return false, err
	}

	// 并发去查询当前用户所有权限
	var wg sync.WaitGroup
	permissionChan := make(chan model.Permission)
	for _, role := range user.Roles {
		wg.Add(1)
		go func(role model.Role) {
			for _, permission := range role.Permissions {
				// api权限控制
				if permission.PermissionType == "api" {
					permissionChan <- permission
				}
			}
			wg.Done()
		}(role)

	}

	go func() {
		wg.Wait()
		close(permissionChan)
	}()

	// 所有有权限的Path
	APIPermission := make([]string, 0)

	// 查询当前登陆用户是否有权调用对应Path资源的权限
	for permission := range permissionChan {
		APIPermission = append(APIPermission, permission.Path)
	}
	for _, b := range APIPermission {
		if b == api {
			return true, nil
		}
	}
	return false, nil
}

func (s *userService) GetMenuTreeByUserAuth(ctx context.Context, userId string) (*v1.GetMenuTreeResponseData, error) {
	user, err := s.userRepo.GetUserWithRolesAndPermission(ctx, userId)
	if err != nil {
		return nil, err
	}

	// Initialize root of the tree
	root := &v1.GetMenuTreeResponseData{
		Label:          "菜单根节点",
		PermissionType: "menu",
		Level:          0,
		Children:       []*v1.GetMenuTreeResponseData{},
	}

	// Map to store pointers to node in the tree
	nodes := map[uint]*v1.GetMenuTreeResponseData{
		0: root,
	}

	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			if permission.PermissionType == "menu" {
				// Create new node
				newNode := &v1.GetMenuTreeResponseData{
					Key:            strconv.FormatUint(uint64(permission.Id), 10),
					ParentId:       permission.ParentId,
					Level:          permission.Level,
					Label:          permission.PermissionName,
					PermissionType: permission.PermissionType,
					Route:          permission.Route,
					RouteFile:      permission.RouteFile,
					CreatedAt:      permission.CreatedAt,
					UpdatedAt:      permission.UpdatedAt,
					DeletedAt:      permission.DeletedAt,
					Children:       []*v1.GetMenuTreeResponseData{},
				}
				// Add this node to its parent's children list
				nodes[permission.ParentId].Children = append(nodes[permission.ParentId].Children, newNode)
				// Add this node to the nodes map
				nodes[permission.Id] = newNode
			}
		}
	}

	return root, nil
}
