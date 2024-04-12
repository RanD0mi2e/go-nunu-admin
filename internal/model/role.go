package model

import (
	"gorm.io/gorm"
	"time"
)

type Role struct {
	gorm.Model
	Id          string `gorm:"primarykey"`
	RoleName    string
	UserId      string
	Permissions []Permission `gorm:"many2many:role_permissions;"`
	deleteFlag  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (m *Role) TableName() string {
	return "role"
}
