package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/miaoming3/photo/internal/config"
	"github.com/miaoming3/photo/internal/model"

	"gorm.io/gorm"
)

// FileService 文件服务接口
type FileService interface {
	Upload(file *model.File, fileContent io.Reader) error
	List(parentID *uint, userID uint) ([]model.File, error)
	Delete(id uint, userID uint) error
	GetByID(id uint, userID uint) (model.File, error)
}

// fileService 文件服务实现
type fileService struct {
	db *gorm.DB
}

// NewFileService 创建文件服务实例
func NewFileService(db *gorm.DB) FileService {
	return &fileService{db}
}

// Upload 上传文件到存储系统并记录元数据
func (s *fileService) Upload(file *model.File, fileContent io.Reader) error {
	// 检查文件类型是否允许
	if !isFileTypeAllowed(file.FileType) {
		return errors.New("不支持的文件类型")
	}

	// 检查文件大小是否超过限制
	if file.FileSize > config.Cfg.Storage.MaxSize*1024*1024 {
		return errors.New("文件大小超过限制")
	}

	// 生成存储路径
	storagePath := config.Cfg.Storage.RootPath
	if file.ParentID != nil {
		// 如果指定了父目录，查询父目录路径
		var parentDir model.File
		if err := s.db.Where("id = ? and is_dir = ?", file.ParentID, true).First(&parentDir).Error; err != nil {
			return errors.New("父目录不存在")
		}
		storagePath = filepath.Join(storagePath, parentDir.FilePath)
	}

	// 确保存储目录存在
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return err
	}

	// 生成文件名
	filename := file.FileName
	if config.Cfg.Storage.HashFilename {
		// 计算文件内容MD5作为文件名
		md5Hash := md5.New()
		if _, err := io.Copy(md5Hash, fileContent); err != nil {
			return err
		}
		file.MD5 = hex.EncodeToString(md5Hash.Sum(nil))
		filename = file.MD5 + filepath.Ext(file.OriginalName)
	}

	// 构建完整文件路径
	filePath := filepath.Join(storagePath, filename)
	relativePath := filepath.Join(filepath.Base(storagePath), filename)
	if file.ParentID != nil {
		var parentDir model.File
		if err := s.db.Where("id = ?", file.ParentID).First(&parentDir).Error; err != nil {
			return fmt.Errorf("failed to find parent directory: %w", err)
		}
		relativePath = filepath.Join(parentDir.FilePath, filename)
	}

	// 保存文件到磁盘
	seeker, ok := fileContent.(io.Seeker)
	if !ok {
		return errors.New("file content does not support seeking")
	}
	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return err
	}
	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, fileContent); err != nil {
		return err
	}

	// 保存文件元数据到数据库
	file.FilePath = relativePath
	return s.db.Create(file).Error
}

// List 获取用户文件列表
func (s *fileService) List(parentID *uint, userID uint) ([]model.File, error) {
	var files []model.File
	query := s.db.Where("user_id = ?", userID)

	if parentID != nil {
		query = query.Where("parent_id = ?", parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	if err := query.Order("is_dir DESC, updated_at DESC").Find(&files).Error; err != nil {
		return nil, err
	}

	return files, nil
}

// Delete 删除用户文件
func (s *fileService) Delete(id uint, userID uint) error {
	// 获取文件信息
	var file model.File
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&file).Error; err != nil {
		return err
	}

	// 如果是目录，检查是否有子文件
	if file.IsDir {
		var childCount int64
		if err := s.db.Model(&model.File{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
			return err
		}
		if childCount > 0 {
			return errors.New("目录不为空，无法删除")
		}
	} else {
		// 删除磁盘上的文件
		fullPath := filepath.Join(config.Cfg.Storage.RootPath, file.FilePath)
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	// 删除数据库记录
	return s.db.Delete(&file).Error
}

// 检查文件类型是否允许
func isFileTypeAllowed(fileType string) bool {
	fmt.Println(config.Cfg.Storage.AllowedTypes)
	for _, allowed := range config.Cfg.Storage.AllowedTypes {
		fmt.Println(allowed)
		if strings.EqualFold(allowed, fileType) {
			return true
		}
	}
	return false
}

func (s *fileService) GetByID(id, userId uint) (file model.File, err error) {
	if err = s.db.Where("id = ? and user_id = ?", id, userId).First(&file).Error; err != nil {
		return
	}
	return
}
