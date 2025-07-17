package model

import (
	"time"
)

type Notification struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	UserID     uint      `gorm:"index:idx_user_id" json:"user_id"`
	Content    string    `gorm:"type:text" json:"content"`
	Type       string    `gorm:"type:varchar(50)" json:"type"` // payment_success, order_status, withdrawal_approved, etc.
	IsRead     bool      `gorm:"default:false" json:"is_read"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	ReadAt     time.Time `gorm:"default:null" json:"read_at,omitempty"`
}

// TableName 设置表名
func (Notification) TableName() string {
	return "notifications"
}