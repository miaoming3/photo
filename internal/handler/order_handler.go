package handler

import (
	"github.com/miaoming3/photo/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/miaoming3/photo/internal/service"
)

// OrderHandler 订单处理器
type OrderHandler struct {
	orderService   service.OrderService
	paymentService *service.PaymentService
}

// NewOrderHandler 创建订单处理器实例
func NewOrderHandler(orderService service.OrderService, paymentService *service.PaymentService) *OrderHandler {
	return &OrderHandler{orderService: orderService, paymentService: paymentService}
}

// CreateOrderRequest 创建订单请求参数
type CreateOrderRequest struct {
	FileID uint `json:"file_id" binding:"required"`
}

// Create 创建订单
func (h *OrderHandler) Create(c *gin.Context) {
	// 获取当前登录用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权访问"})
		return
	}

	// 绑定请求参数
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 创建订单
	order, err := h.orderService.CreateOrder(userID.(uint), req.FileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 生成支付链接
	payURL, err := h.paymentService.CreatePayment(order.OrderNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成支付链接失败: " + err.Error()})
		return
	}

	// 返回订单信息和支付链接
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "订单创建成功",
		"data": gin.H{
			"order":   order,
			"pay_url": payURL,
		},
	})
}

// GetOrderStatusRequest 查询订单状态请求参数
type GetOrderStatusRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
}

// List 查询订单列表
func (h *OrderHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	var orders []model.Order
	var total int64

	total, _ = h.orderService.FindByCount(userID)
	orders, _ = h.orderService.Pagetions(userID, offset, pageSize, "created_at desc")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":       orders,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetStatus 查询订单状态
func (h *OrderHandler) GetStatus(c *gin.Context) {
	// 获取当前登录用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权访问"})
		return
	}

	// 绑定请求参数
	var req GetOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 查询订单
	order, err := h.orderService.GetByUserOrder(req.OrderNo, userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	// 返回订单状态
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"order_no": order.OrderNo,
			"status":   order.Status,
			"amount":   order.Amount,
		},
	})
}
