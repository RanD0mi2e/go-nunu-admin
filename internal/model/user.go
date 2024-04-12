package model

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	Id        uint   `gorm:"primarykey"`
	UserId    string `gorm:"unique;not null"`
	Nickname  string `gorm:"not null"`
	Password  string `gorm:"not null"`
	Email     string `gorm:"not null"`
	Roles     []Role `gorm:"many2many:user_role"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *User) TableName() string {
	return "users"
}
