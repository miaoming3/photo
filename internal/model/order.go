package model

import (
	"time"

	"gorm.io/gorm"
)

// Order 订单模型
 type Order struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	OrderNo        string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"order_no"` // 订单编号
	UserID         uint           `gorm:"index;not null" json:"user_id"` // 购买用户ID
	FileID         uint           `gorm:"index;not null" json:"file_id"` // 文件ID
	Amount         float64        `gorm:"type:decimal(10,2);not null" json:"amount"` // 订单金额
	Status         int            `gorm:"type:int;default:0" json:"status"` // 订单状态：0-待支付，1-已支付，2-已取消
	UserIncome     float64        `gorm:"type:decimal(10,2);default:0" json:"user_income"` // 用户收益(70%)
	PlatformIncome float64        `gorm:"type:decimal(10,2);default:0" json:"platform_income"` // 平台收益(30%)
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	File           File           `gorm:"foreignKey:FileID" json:"file,omitempty"` // 关联文件
	User           User           `gorm:"foreignKey:UserID" json:"user,omitempty"` // 关联用户
}

// TableName 设置表名
func (Order) TableName() string {
	return "orders"
}