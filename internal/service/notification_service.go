package service

import (
	"context"
	"github.com/miaoming3/photo/internal/model"
	"time"

	"gorm.io/gorm"
)

type NotificationService struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// CreateNotification 创建新通知
func (s *NotificationService) CreateNotification(ctx context.Context, userID uint, content string, notificationType string) error {
	notification := &model.Notification{
		UserID:    userID,
		Content:   content,
		Type:      notificationType,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	return s.db.WithContext(ctx).Create(notification).Error
}

// GetUserNotifications 获取用户通知列表
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uint, page, pageSize int) ([]model.Notification, int64, error) {
	var notifications []model.Notification
	var total int64

	offset := (page - 1) * pageSize

	if err := s.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// MarkAsRead 将通知标记为已读
func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID uint) error {
	return s.db.WithContext(ctx).Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": time.Now(),
		}).Error
}

// GetUnreadCount 获取未读通知数量
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	return count, s.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
}
