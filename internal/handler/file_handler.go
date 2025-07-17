package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/miaoming3/photo/internal/model"
	"github.com/miaoming3/photo/internal/service"
)

// FileHandler 文件处理器
type FileHandler struct {
	fileService  service.FileService
	orderService service.OrderService
}

// Get 获取文件详情
func (h *FileHandler) Get(c *gin.Context) {
	// 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的文件ID"})
		return
	}

	// 获取当前登录用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权访问"})
		return
	}

	// 查询文件
	file, err := h.fileService.GetByID(uint(id), userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	// 付费文件权限检查
	if file.IsPaid && file.UserID != userID.(uint) {
		// 查询用户是否购买该文件
		if _, err := h.orderService.IsPay(file.ID, userID.(uint), 1); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "该文件为付费资源，请先购买"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    file,
	})
}

// NewFileHandler 创建文件处理器实例
func NewFileHandler(fileService service.FileService, orderService service.OrderService) *FileHandler {
	return &FileHandler{fileService: fileService, orderService: orderService}
}

// Upload 文件上传接口
// @Summary 上传文件
// @Description 上传图片文件到图床
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Param parent_id formData int false "父目录ID"
// @Success 200 {object} gin.H{"code":int,"data":model.File,"message":string}
// @Failure 400 {object} gin.H{"code":int,"message":string}
// @Router /api/upload [post]
func (h *FileHandler) Upload(c *gin.Context) {
	// 获取上传文件
	// 获取当前登录用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证的用户"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "获取文件失败: " + err.Error()})
		return
	}
	defer file.Close()

	// 获取父目录ID
	parentIDStr := c.PostForm("parent_id")
	var parentID *uint
	if parentIDStr != "" {
		pid, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的父目录ID"})
			return
		}
		uid := uint(pid)
		parentID = &uid
	}

	// 获取是否收费
	isPaidStr := c.PostForm("is_paid")
	isPaid := false
	if isPaidStr == "true" {
		isPaid = true
	}

	// 获取价格
	priceStr := c.PostForm("price")
	price := 0.0
	if priceStr != "" {
		parsedPrice, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的价格"})
			return
		}
		price = parsedPrice
	}

	// 获取文件信息
	fileSize := header.Size
	fileName := header.Filename
	fileType := header.Header.Get("Content-Type")

	// 创建文件模型
	fileModel := &model.File{
		OriginalName: fileName,
		FileName:     fileName,
		FileType:     fileType,
		FileSize:     fileSize,
		IsPaid:       isPaid,
		Price:        price,
		ParentID:     parentID,
		UserID:       userID.(uint),
		IsDir:        false,
	}

	// 调用服务层上传文件
	if err := h.fileService.Upload(fileModel, file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "上传失败: " + err.Error()})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "上传成功",
		"data":    fileModel,
	})
}

// List 文件列表接口
// @Summary 获取文件列表
// @Description 获取指定目录下的文件列表
// @Tags files
// @Accept json
// @Produce json
// @Param parent_id query int false "父目录ID"
// @Success 200 {object} gin.H{"code":int,"data":[]model.File,"message":string}
// @Failure 400 {object} gin.H{"code":int,"message":string}
// @Router /api/files [get]
func (h *FileHandler) List(c *gin.Context) {
	// 获取父目录ID
	parentIDStr := c.Query("parent_id")
	var parentID *uint
	if parentIDStr != "" {
		pid, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的父目录ID"})
			return
		}
		uid := uint(pid)
		parentID = &uid
	}

	// 获取当前登录用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证的用户"})
		return
	}

	// 调用服务层获取文件列表
	files, err := h.fileService.List(parentID, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取文件列表失败: " + err.Error()})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    files,
	})
}

// Delete 文件删除接口
// @Summary 删除文件
// @Description 根据ID删除文件或目录
// @Tags files
// @Accept json
// @Produce json
// @Param id path int true "文件ID"
// @Success 200 {object} gin.H{"code":int,"message":string}
// @Failure 400 {object} gin.H{"code":int,"message":string}
// @Router /api/files/{id} [delete]
func (h *FileHandler) Delete(c *gin.Context) {
	// 获取文件ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的文件ID"})
		return
	}

	// 获取当前登录用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证的用户"})
		return
	}

	// 调用服务层删除文件
	if err := h.fileService.Delete(uint(id), userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "删除失败: " + err.Error()})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}
