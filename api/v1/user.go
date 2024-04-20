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
	Key            string                     `json:"key"`
	Label          string                     `json:"label"`
	Sort           string                     `json:"sort"`
	PermissionType string                     `json:"permission_type"`
	ParentId       string                     `json:"parent_id"`
	Level          uint                       `json:"level"`
	Icon           string                     `json:"icon"`
	Route          string                     `json:"route"`
	RouteFile      string                     `json:"route_file"`
	Path           string                     `json:"path"`
	Method         string                     `json:"method"`
	Children       []*GetMenuTreeResponseData `json:"children,omitempty"`
	CreatedAt      string                     `json:"created_at"`
	UpdatedAt      string                     `json:"updated_at"`
	DeletedAt      gorm.DeletedAt             `json:"deleted_at"`
}

type GetMenuTreeResponse struct {
	Response
	Data GetMenuTreeResponseData
}
