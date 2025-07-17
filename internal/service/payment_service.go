package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/miaoming3/photo/internal/model"

	"gorm.io/gorm"
)

// PaymentGateway 支付网关接口
type PaymentGateway interface {
	CreatePayment(order *model.Order) (string, error) // 创建支付链接
	VerifyPayment(orderNo string) (bool, error)       // 验证支付状态
}

// MockPaymentGateway 模拟支付网关实现
type MockPaymentGateway struct{}

// CreatePayment 生成模拟支付链接
func (m *MockPaymentGateway) CreatePayment(order *model.Order) (string, error) {
	// 实际项目中这里会调用真实支付网关API
	return fmt.Sprintf("https://mock-payment.com/pay?order_no=%s&amount=%.2f", order.OrderNo, order.Amount), nil
}

// VerifyPayment 模拟验证支付状态
func (m *MockPaymentGateway) VerifyPayment(orderNo string) (bool, error) {
	// 实际项目中这里会调用支付网关查询接口
	// 模拟支付成功（实际应根据支付网关返回结果判断）
	return true, nil
}

// PaymentService 支付服务
type PaymentService struct {
	db             *gorm.DB
	paymentGateway PaymentGateway
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(db *gorm.DB, gateway PaymentGateway) *PaymentService {
	return &PaymentService{
		db:             db,
		paymentGateway: gateway,
	}
}

// CreatePayment 创建支付链接
func (s *PaymentService) CreatePayment(orderNo string) (string, error) {
	// 查询订单
	var order model.Order
	if err := s.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return "", errors.New("订单不存在")
	}

	// 检查订单状态
	if order.Status != 0 {
		return "", errors.New("订单状态异常")
	}

	// 调用支付网关生成支付链接
	payURL, err := s.paymentGateway.CreatePayment(&order)
	if err != nil {
		return "", fmt.Errorf("生成支付链接失败: %w", err)
	}

	return payURL, nil
}

// HandlePaymentCallback 处理支付回调
func (s *PaymentService) HandlePaymentCallback(orderNo string) error {
	// 验证支付状态
	paid, err := s.paymentGateway.VerifyPayment(orderNo)
	if err != nil {
		return fmt.Errorf("验证支付失败: %w", err)
	}

	if !paid {
		return errors.New("支付未完成")
	}

	// 更新订单状态
	return s.db.Transaction(func(tx *gorm.DB) error {
		var order model.Order
		if err := tx.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}

		if order.Status == 1 {
			return errors.New("订单已支付")
		}

		// 更新订单状态为已支付
		if err := tx.Model(&order).Update("status", 1).Update("updated_at", time.Now()).Error; err != nil {
			return err
		}

		// 更新文件所有者余额
		return tx.Model(&model.User{}).Where("id = ?", order.FileID).
			Update("balance", gorm.Expr("balance + ?", order.UserIncome)).Error
	})
}
