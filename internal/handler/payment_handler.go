package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miaoming3/photo/internal/service"
)

// PaymentHandler 支付处理器
 type PaymentHandler struct {
	paymentService *service.PaymentService
}

// NewPaymentHandler 创建支付处理器实例
func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// HandleCallback 处理支付回调
func (h *PaymentHandler) HandleCallback(c *gin.Context) {
	// 从回调参数中获取订单号（实际项目中需根据支付网关的回调格式解析）
	orderNo := c.PostForm("order_no")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少订单号"})
		return
	}

	// 处理支付结果
	if err := h.paymentService.HandlePaymentCallback(orderNo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "处理支付回调失败: " + err.Error()})
		return
	}

	// 返回支付网关要求的成功响应格式（通常是特定格式的字符串，如"success"）
	c.String(http.StatusOK, "success")
}