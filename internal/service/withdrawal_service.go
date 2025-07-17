package service

import (
	"errors"
	"github.com/miaoming3/photo/internal/model"

	"gorm.io/gorm"
)

// WithdrawalService 提现服务接口
type WithdrawalService interface {
	CreateWithdrawal(req *model.Withdrawal) error
	GetUserWithdrawals(userID uint) ([]model.Withdrawal, error)
}

// withdrawalService 提现服务实现
type withdrawalService struct {
	db *gorm.DB
}

// NewWithdrawalService 创建提现服务实例
func NewWithdrawalService(db *gorm.DB) WithdrawalService {
	return &withdrawalService{db}
}

// CreateWithdrawal 创建提现申请
func (s *withdrawalService) CreateWithdrawal(req *model.Withdrawal) error {
	// 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 查询用户余额
		var user model.User
		if err := tx.Where("id = ?", req.UserID).First(&user).Error; err != nil {
			return errors.New("用户不存在")
		}

		// 检查余额是否充足
		if user.Balance < req.Amount {
			return errors.New("余额不足")
		}

		// 创建提现记录
		if err := tx.Create(req).Error; err != nil {
			return err
		}

		// 冻结用户余额（实际项目中可能需要单独的冻结字段）
		return tx.Model(&user).Update("balance", gorm.Expr("balance - ?", req.Amount)).Error
	})
}

// GetUserWithdrawals 查询用户提现记录
func (s *withdrawalService) GetUserWithdrawals(userID uint) ([]model.Withdrawal, error) {
	var withdrawals []model.Withdrawal
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&withdrawals).Error; err != nil {
		return nil, err
	}
	return withdrawals, nil
}
