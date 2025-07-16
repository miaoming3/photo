package models

import (
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB

// User 用户模型
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:50" json:"username"`
	Password  string         `gorm:"size:100" json:"-"`
	Phone     string         `gorm:"size:20" json:"phone"`
	Avatar    string         `json:"avatar"`
	Balance   float64        `gorm:"default:0" json:"balance"`
	IsAgent   bool           `gorm:"default:false" json:"is_agent"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Photo 照片模型
type Photo struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index" json:"user_id"`
	URL       string         `json:"url"`
	QRCodeURL string         `json:"qr_code_url"`
	Price     float64        `json:"price"`
	Visible   bool           `gorm:"default:true" json:"visible"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// AgentWithdrawal 代理提现模型
type AgentWithdrawal struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	AgentID   uint           `gorm:"index" json:"agent_id"`
	Amount    float64        `json:"amount"`
	Status    string         `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, approved, rejected
	BankInfo  string         `json:"bank_info"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Order 支付订单模型
type Order struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	OrderID   string         `gorm:"uniqueIndex;size:50" json:"order_id"` // 外部订单号
	UserID    uint           `gorm:"index" json:"user_id"`
	Amount    float64        `json:"amount"`
	Status    string         `gorm:"type:tinyint(1);default:'1'" json:"status"` // 1待支付, 2已支付, 3已取消
	PhotoID   uint           `gorm:"index;default:0" json:"photo_id"`           // 关联的照片ID
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
