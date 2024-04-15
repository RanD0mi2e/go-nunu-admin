package model

import (
	"gorm.io/gorm"
	"time"
)

type Role struct {
	Id          uint   `gorm:"primarykey;auto_increment"`
	RoleLabel   string `gorm:"unique"`
	RoleName    string
	UserId      string
	Permissions []Permission `gorm:"many2many:role_permissions;"`
	DeleteFlag  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (m *Role) TableName() string {
	return "role"
}
