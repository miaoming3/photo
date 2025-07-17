package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Username     string         `gorm:"type:varchar(50);not null;unique" json:"username"`
	Email        string         `gorm:"type:varchar(191);not null;unique" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Balance      float64        `gorm:"type:decimal(10,2);default:0" json:"balance"` // 用户余额
	Nickname     string         `gorm:"type:varchar(255)" json:"nickname"`
	Phone        string         `gorm:"type:varchar(20);unique" json:"phone"`
	Avatar       string         `json:"avatar"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}
