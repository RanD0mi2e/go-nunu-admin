package service

import (
	v1 "admin-webrtc-go/api/v1"
	"admin-webrtc-go/internal/model"
	"admin-webrtc-go/internal/repository"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	sortPkg "sort"
	"strconv"
	"time"
)

type UserService interface {
	Register(ctx context.Context, req *v1.RegisterRequest) error
	Login(ctx context.Context, req *v1.LoginRequest) (string, error)
	GetProfile(ctx context.Context, userId string) (*v1.GetProfileResponseData, error)
	UpdateProfile(ctx context.Context, userId string, req *v1.UpdateProfileRequest) error
	CheckAPIAuthPermission(ctx context.Context, userId string, api string) (bool, error)
	GetMenuTreeByUserAuth(ctx context.Context, userId string, sort string) ([]*v1.GetMenuTreeResponseData, error)
}

func NewUserService(service *Service, userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
		Service:  service,
	}
}

type treeNode struct {
	repository.LoginedUser
	Children []*treeNode
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
	users, err := s.userRepo.GetUserWithRolesAndPermission(ctx, userId, "api", "")
	// 数据库查不到用户的权限，返回false
	if err != nil {
		return false, err
	}
	for _, user := range *users {
		if user.Path == api {
			return true, nil
		}
	}

	return false, nil
}

func (s *userService) GetMenuTreeByUserAuth(ctx context.Context, userId string, sort string) ([]*v1.GetMenuTreeResponseData, error) {
	users, err := s.userRepo.GetUserWithRolesAndPermission(ctx, userId, "menu", sort)
	if err != nil {
		return nil, err
	}

	menuTree := convertToTree(users)
	if sort == "asc" || sort == "desc" {
		recursiveSort(menuTree, sort)
	}

	if len(menuTree) == 0 {
		return nil, errors.New("menuTree is empty")
	}

	return menuTree, nil

}

func convertToTree(users *[]repository.LoginedUser) []*v1.GetMenuTreeResponseData {
	nodeMap := make(map[string]*v1.GetMenuTreeResponseData)

	for _, user := range *users {
		node := &v1.GetMenuTreeResponseData{
			Key:       strconv.FormatUint(uint64(user.PermissionId), 10),
			Label:     user.PermissionName,
			Sort:      user.Sort,
			ParentId:  strconv.FormatUint(uint64(user.ParentId), 10),
			Level:     user.Level,
			Route:     user.Route,
			RouteFile: user.RouteFile,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Children:  []*v1.GetMenuTreeResponseData{},
		}
		nodeMap[strconv.FormatUint(uint64(user.PermissionId), 10)] = node
	}

	var root []*v1.GetMenuTreeResponseData
	for _, node := range nodeMap {
		if parent, exist := nodeMap[node.ParentId]; exist {
			parent.Children = append(parent.Children, node)
		} else {
			root = append(root, node)
		}
	}

	return root
}

func recursiveSort(nodes []*v1.GetMenuTreeResponseData, sort string) {
	if sort == "asc" {
		sortPkg.Slice(nodes, func(i, j int) bool {
			return nodes[i].Sort < nodes[j].Sort
		})
	}
	if sort == "desc" {
		sortPkg.Slice(nodes, func(i, j int) bool {
			return nodes[i].Sort > nodes[j].Sort
		})
	}

	for _, node := range nodes {
		recursiveSort(node.Children, sort)
	}
}
