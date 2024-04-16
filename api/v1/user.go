package v1

import "gorm.io/gorm"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"1234@gmail.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"1234@gmail.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}
type LoginResponseData struct {
	AccessToken string `json:"accessToken"`
}
type LoginResponse struct {
	Response
	Data LoginResponseData
}

type UpdateProfileRequest struct {
	Nickname string `json:"nickname" example:"alan"`
	Email    string `json:"email" binding:"required,email" example:"1234@gmail.com"`
}
type GetProfileResponseData struct {
	UserId   string `json:"userId"`
	Nickname string `json:"nickname" example:"alan"`
}
type GetProfileResponse struct {
	Response
	Data GetProfileResponseData
}
type GetMenuTreeResponseData struct {
	ID             uint                       `json:"id"`
	PermissionName string                     `json:"permissionName"`
	PermissionType string                     `json:"permissionType"`
	ParentId       uint                       `json:"parentId"`
	Level          int                        `json:"level"`
	Icon           string                     `json:"icon"`
	Route          string                     `json:"route"`
	RouteFile      string                     `json:"routeFile"`
	Path           string                     `json:"path"`
	Method         string                     `json:"method"`
	Children       []*GetMenuTreeResponseData `json:"children"`
	CreatedAt      string                     `json:"createdAt"`
	UpdatedAt      string                     `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt             `json:"deletedAt"`
}

type GetMenuTreeResponse struct {
	Response
	Data GetMenuTreeResponseData
}
