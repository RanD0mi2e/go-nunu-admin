package service

import (
	v1 "admin-webrtc-go/api/v1"
	"admin-webrtc-go/internal/model"
	"admin-webrtc-go/internal/repository"
	"context"
	"golang.org/x/crypto/bcrypt"
	sortPkg "sort"
	"strconv"
	"sync"
	"time"
)

type UserService interface {
	Register(ctx context.Context, req *v1.RegisterRequest) error
	Login(ctx context.Context, req *v1.LoginRequest) (string, error)
	GetProfile(ctx context.Context, userId string) (*v1.GetProfileResponseData, error)
	UpdateProfile(ctx context.Context, userId string, req *v1.UpdateProfileRequest) error
	CheckAPIAuthPermission(ctx context.Context, userId string, api string) (bool, error)
	GetMenuTreeByUserAuth(ctx context.Context, userId string, sort string) (*v1.GetMenuTreeResponseData, error)
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
	user, err := s.userRepo.GetUserWithRolesAndPermission(ctx, userId, "")
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

func (s *userService) GetMenuTreeByUserAuth(ctx context.Context, userId string, sort string) (*v1.GetMenuTreeResponseData, error) {
	user, err := s.userRepo.GetUserWithRolesAndPermission(ctx, userId, sort)
	if err != nil {
		return nil, err
	}

	// Initialize root of the tree
	root := &v1.GetMenuTreeResponseData{
		Label:          "菜单根节点",
		PermissionType: "menu",
		Level:          0,
		Key:            "0",
		Children:       []*v1.GetMenuTreeResponseData{},
	}

	var rootNodes []*v1.GetMenuTreeResponseData

	// 节点暂存表，暂时找不到父节点的节点先放进表里
	parentIdMap := make(map[string][]*v1.GetMenuTreeResponseData)

	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			if permission.PermissionType == "menu" {
				permissionIdStr := strconv.FormatUint(uint64(permission.Id), 10)
				parentIdStr := strconv.FormatUint(uint64(permission.ParentId), 10)
				// Create new node
				newNode := &v1.GetMenuTreeResponseData{
					Key:            permissionIdStr,
					ParentId:       parentIdStr,
					Sort:           permission.Sort,
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
				if permission.Level == 1 {
					rootNodes = append(rootNodes, newNode)
				} else {
					parentIdMap[parentIdStr] = append(parentIdMap[parentIdStr], newNode)
				}
				// Add this node to its parent's children list
				//nodes[permission.ParentId].Children = append(nodes[permission.ParentId].Children, newNode)
				// Add this node to the nodes map
				//nodes[permission.Key] = newNode
			}
		}
	}
	// 最顶层父层级排序
	if sort == "asc" {
		sortPkg.Slice(rootNodes, func(i, j int) bool {
			return rootNodes[i].Level < rootNodes[j].Level
		})
	}
	if sort == "desc" {
		sortPkg.Slice(rootNodes, func(i, j int) bool {
			return rootNodes[i].Level > rootNodes[j].Level
		})
	}
	// Build the tree structure by connecting parent and child nodes
	for _, node := range rootNodes {
		parentId := node.Key
		if children, ok := parentIdMap[parentId]; ok {
			// 子层级排序后再添加到父层级中
			if sort == "asc" {
				sortPkg.Slice(children, func(i, j int) bool {
					return children[i].Level < children[j].Level
				})
			}
			if sort == "desc" {
				sortPkg.Slice(children, func(i, j int) bool {
					return children[i].Level > children[j].Level
				})
			}
			node.Children = children
		}
	}

	root.Children = rootNodes

	// Return the root node of the tree
	return root, nil
}
