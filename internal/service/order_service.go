package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/miaoming3/photo/internal/model"

	"gorm.io/gorm"
)

// OrderService 订单服务接口
type OrderService interface {
	CreateOrder(userID, fileID uint) (*model.Order, error)
	UpdateOrderStatus(orderNo string, status int) error
	IsPay(id, userID, status uint) (model.Order, error)
	Lists(userID uint, orderBy string) ([]model.Order, error)
	GetByUserOrder(orderSn string, userID uint) (model.Order, error)
	FindByCount(userID uint) (total int64, err error)
	Pagetions(userID uint, page, size int, orderBy string) ([]model.Order, error)
}

// orderService 订单服务实现
type orderService struct {
	db *gorm.DB
}

// NewOrderService 创建订单服务实例
func NewOrderService(db *gorm.DB) OrderService {
	return &orderService{db}
}

// CreateOrder 创建订单
func (s *orderService) CreateOrder(userID, fileID uint) (*model.Order, error) {
	// 查询文件信息
	var file model.File
	if err := s.db.Where("id = ? AND is_paid = ?", fileID, true).First(&file).Error; err != nil {
		return nil, errors.New("文件不存在或为免费文件")
	}

	// 生成订单号
	orderNo := generateOrderNo()

	// 计算收益分成 (用户70%，平台30%)
	userIncome := file.Price * 0.7
	platformIncome := file.Price * 0.3

	// 创建订单
	order := &model.Order{
		OrderNo:        orderNo,
		UserID:         userID,
		FileID:         fileID,
		Amount:         file.Price,
		Status:         0, // 待支付
		UserIncome:     userIncome,
		PlatformIncome: platformIncome,
	}

	// 使用事务创建订单
	return order, s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 订单创建成功后逻辑（实际项目中这里可能需要调用支付接口）
		return nil
	})
}

// UpdateOrderStatus 更新订单状态
func (s *orderService) UpdateOrderStatus(orderNo string, status int) error {
	// 查询订单
	var order model.Order
	if err := s.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}

	// 如果订单已支付，不再处理
	if order.Status == 1 {
		return errors.New("订单已支付")
	}

	// 使用事务更新订单状态和用户余额
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 更新订单状态
		if err := tx.Model(&order).Update("status", status).Error; err != nil {
			return err
		}

		// 如果是支付成功状态，更新文件所有者余额
		if status == 1 {
			// 查询文件获取所有者ID
			var file model.File
			if err := tx.Where("id = ?", order.FileID).First(&file).Error; err != nil {
				return err
			}

			// 更新所有者余额
			return tx.Model(&model.User{}).Where("id = ?", file.UserID).
				Update("balance", gorm.Expr("balance + ?", order.UserIncome)).Error
		}

		return nil
	})
}

// generateOrderNo 生成唯一订单号
func generateOrderNo() string {
	// 格式: 年月日时分秒 + 3位随机数
	timeStr := time.Now().Format("20060102150405")
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(900) + 100 // 100-999的随机数
	return fmt.Sprintf("%s%d", timeStr, random)
}

func (s *orderService) IsPay(id, userID, status uint) (order model.Order, err error) {
	if err = s.db.Where("file_id = ? AND user_id = ? AND status = ?", id, userID, status).
		First(&order).Error; err != nil {
		return
	}
	return
}

func (s *orderService) Lists(userID uint, orderBy string) (orders []model.Order, err error) {
	if orderBy == "" {
		orderBy = string("created_at DESC")
	}
	if err = s.db.Where("user_id = ?", userID).Order(orderBy).Find(&orders).Error; err != nil {
		return
	}
	return
}

func (s *orderService) GetByUserOrder(order_sn string, userID uint) (order model.Order, err error) {
	if err := s.db.Where("order_sn = ? and user_id = ?", order_sn, userID).First(&order).Error; err != nil {
		return model.Order{}, err
	}
	return
}

func (s *orderService) FindByCount(userID uint) (total int64, err error) {
	if err = s.db.Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return
	}
	return
}
func (s *orderService) Pagetions(userID uint, page, size int, orderBy string) (orders []model.Order, err error) {
	if err = s.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(size).Offset(page).Find(&orders).Error; err != nil {
		return nil, err
	}
	return

}
