package model

import (
	"time"

	"gorm.io/gorm"
)

// File 图片文件模型
type File struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	FileName     string         `gorm:"type:varchar(255);not null" json:"file_name"`
	OriginalName string         `gorm:"type:varchar(255);not null" json:"original_name"`
	FilePath     string         `gorm:"type:varchar(512);not null;unique" json:"file_path"`
	FileType     string         `gorm:"type:varchar(50);not null" json:"file_type"`
	FileSize     int64          `gorm:"type:int;not null" json:"file_size"`
	IsPaid       bool           `gorm:"default:false" json:"is_paid"` // 是否收费
	Price        float64        `gorm:"default:0" json:"price"`       // 价格(元)
	MD5          string         `gorm:"type:char(32);uniqueIndex" json:"md5"`
	IsDir        bool           `gorm:"type:bool;default:false" json:"is_dir"`
	ParentID     *uint          `gorm:"index" json:"parent_id,omitempty"`
	UserID       uint           `gorm:"index" json:"user_id"` // 文件所属用户ID
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName 设置表名
func (File) TableName() string {
	return "files"
}
