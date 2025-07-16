package photo

import (
	"errors"
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/miaomimg3/photo/models"
	"github.com/miaomimg3/photo/payment"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"io"
	"os"
	"path/filepath"
	"github.com/google/uuid"
)

type PhotoHandler struct{}

// Upload 上传照片（需微信支付）
func (h *PhotoHandler) Upload(c *gin.Context) {
	var req struct {
		UserID uint    `json:"user_id" binding:"required"`
		Price  float64 `json:"price" binding:"required,min=0.01"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 创建微信支付订单
	wechatPay := payment.NewWechatPay()
	order, err := wechatPay.CreateOrder(req.UserID, req.Price, "照片上传")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建支付订单失败"})
		return
	}

	// 2. 返回支付二维码URL
	c.JSON(http.StatusOK, gin.H{
		"order_id": order.OrderID,
		"qr_code_url": order.QRCodeURL,
		"pay_amount": req.Price,
	})

	// 3. 保存订单到数据库
	order := models.Order{
		OrderID: order.OrderID,
		UserID:  req.UserID,
		Amount:  req.Price,
		Status:  "pending",
	}
	if err := models.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存订单失败"})
		return
	}

	// 4. 处理文件上传
	file, _, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传照片文件"})
		return
	}
	defer file.Close()

	// 创建存储目录
	uploadDir := filepath.Join("uploads", time.Now().Format("20060102"))
	os.MkdirAll(uploadDir, 0755)

	// 生成唯一文件名
	filename := uuid.New().String() + filepath.Ext(file.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// 保存文件
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "复制文件失败"})
		return
	}

	// 5. 创建照片记录（待支付）
	photo := models.Photo{
		UserID:  req.UserID,
		URL:     filePath,
		Price:   req.Price,
		Visible: false,
	}
	if err := models.DB.Create(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建照片记录失败"})
		return
	}

	// 6. 更新订单关联的照片ID
	models.DB.Model(&order).Update("photo_id", photo.ID)

	// 7. 返回支付信息
	c.JSON(http.StatusOK, gin.H{
		"order_id":    order.OrderID,
		"qr_code_url": order.QRCodeURL,
		"pay_amount":  req.Price,
	})
}

// List 获取公开照片列表
func (h *PhotoHandler) List(c *gin.Context) {
	var photos []models.Photo
	if err := models.DB.Where("visible = ?", true).Find(&photos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取照片失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"photos": photos})
}

// Get 获取单张照片详情
func (h *PhotoHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var photo models.Photo
	if err := models.DB.Where("id = ? AND visible = ?", id, true).First(&photo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "照片不存在或未公开"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取照片失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"photo": photo})
}

// AdminList 管理员获取所有照片
func (h *PhotoHandler) AdminList(c *gin.Context) {
	var photos []models.Photo
	if err := models.DB.Find(&photos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取照片失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"photos": photos})
}

// UpdateVisible 更新照片可见性
func (h *PhotoHandler) UpdateVisible(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Visible bool `json:"visible" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := models.DB.Model(&models.Photo{}).Where("id = ?", id).Update("visible", req.Visible).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新照片状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "照片状态已更新"})
}

// HandlePaymentNotify 处理微信支付回调
func (h *PhotoHandler) HandlePaymentNotify(c *gin.Context) {
	// 读取回调数据
	body, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusBadRequest, "读取回调数据失败")
		return
	}

	// 处理支付回调
	wechatPay := payment.NewWechatPay()
	success, err := wechatPay.HandleNotify(body)
	if err != nil || !success {
		// 微信支付要求必须返回SUCCESS或FAIL
		c.String(http.StatusOK, `<xml><return_code><![CDATA[FAIL]]></return_code><return_msg><![CDATA[%s]]></return_msg></xml>`, err.Error())
		return
	}

	// 解析回调数据获取订单信息
	var notifyData struct {
		OutTradeNo string `json:"out_trade_no"`
		TradeState string `json:"trade_state"`
	}
	if err := json.Unmarshal(body, &notifyData); err != nil {
		c.String(http.StatusOK, `<xml><return_code><![CDATA[FAIL]]></return_code><return_msg><![CDATA[解析数据失败]]></return_msg></xml>`)
		return
	}

	// 查询订单
	var order models.Order
	if err := models.DB.Where("order_id = ?", notifyData.OutTradeNo).First(&order).Error; err != nil {
		c.String(http.StatusOK, `<xml><return_code><![CDATA[FAIL]]></return_code><return_msg><![CDATA[获取订单失败]]></return_msg></xml>`)
		return
	}

	// 如果订单已处理，直接返回成功
	if order.Status == "paid" {
		c.String(http.StatusOK, `<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>`)
		return
	}

	// 更新订单状态
	order.Status = "paid"
	models.DB.Save(&order)

	// 更新照片可见性
	models.DB.Model(&models.Photo{}).Where("id = ?", order.PhotoID).Update("visible", true)

	// 生成照片二维码
	qrCodePath := generateQRCode(order.PhotoID)
	models.DB.Model(&models.Photo{}).Where("id = ?", order.PhotoID).Update("qr_code_url", qrCodePath)

	// 计算代理佣金
	h.calculateAgentCommission(order.UserID, order.Amount)

	// 返回成功响应给微信支付
	c.String(http.StatusOK, `<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>`)
}

// 生成二维码
func generateQRCode(photoID uint) string {
	// 实现二维码生成逻辑
	// ...
}

// 计算代理佣金
func (h *PhotoHandler) calculateAgentCommission(userID uint, amount float64) {
	// 实现代理佣金计算逻辑
	// ...
}