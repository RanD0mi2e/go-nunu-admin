package model

import (
	"gorm.io/gorm"
)

type Permission struct {
	Id             uint          `gorm:"primarykey;auto_increment"`
	PermissionName string        // 权限名称
	PermissionType string        // 权限类型
	ParentId       uint          // 父节点id
	Level          int           // 层级
	Icon           string        // 菜单图标
	Route          string        // 路由地址
	RouteFile      string        // 路由地址对应前端文件
	Path           string        // 有权访问的路径
	Method         string        // 有权访问的方法
	Children       []*Permission `gorm:"foreignKey:ParentId;references:ParentId"`
	CreatedAt      string
	UpdatedAt      string
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (m *Permission) TableName() string {
	return "permission"
}
