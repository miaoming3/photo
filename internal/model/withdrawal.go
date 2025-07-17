package model

import (
	"time"

	"gorm.io/gorm"
)

// Withdrawal 提现记录模型
type Withdrawal struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	UserID        uint           `gorm:"index;not null" json:"user_id"`                   // 用户ID
	Amount        float64        `gorm:"type:decimal(10,2);not null" json:"amount"`       // 提现金额
	Status        int            `gorm:"type:int;default:0" json:"status"`                // 状态：0-待审核，1-已批准，2-已拒绝，3-已完成
	BankAccount   string         `gorm:"type:varchar(100);not null" json:"bank_account"`  // 银行账号
	BankName      string         `gorm:"type:varchar(50);not null" json:"bank_name"`      // 银行名称
	AccountHolder string         `gorm:"type:varchar(50);not null" json:"account_holder"` // 账户持有人
	Remark        string         `gorm:"type:text" json:"remark"`                         // 备注
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"` // 关联用户
	ReviewerID    uint           `gorm:"default:0" json:"reviewer_id,omitempty"`
	ReviewTime    int64          `gorm:"default:0" json:"review_time,omitempty"`
}

// TableName 设置表名
func (Withdrawal) TableName() string {
	return "withdrawals"
}
