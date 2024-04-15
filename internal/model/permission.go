package model

import (
	"gorm.io/gorm"
	"time"
)

type Permission struct {
	gorm.Model
	Id             uint `gorm:"primarykey;auto_increment"`
	PermissionName string
	PermissionType string        // 权限类型
	ParentId       uint          // 父节点id
	Level          int           // 层级
	Icon           string        // 菜单图标
	Route          string        // 路由地址
	RouteFile      string        // 路由地址对应前端文件
	Path           string        // 有权访问的路径
	Method         string        // 有权访问的方法
	Children       []*Permission `gorm:"foreignKey:ParentId;references:ParentId"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (m *Permission) TableName() string {
	return "permission"
}
