package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miaoming3/photo/internal/model"
	"github.com/miaoming3/photo/internal/service"
)

// WithdrawalHandler 提现处理器
 type WithdrawalHandler struct {
	withdrawalService service.WithdrawalService
}

// NewWithdrawalHandler 创建提现处理器实例
func NewWithdrawalHandler(withdrawalService service.WithdrawalService) *WithdrawalHandler {
	return &WithdrawalHandler{withdrawalService: withdrawalService}
}

// CreateWithdrawalRequest 提现申请请求参数
 type CreateWithdrawalRequest struct {
	Amount        float64 `json:"amount" binding:"required,min=0.01"`
	BankAccount   string  `json:"bank_account" binding:"required"`
	BankName      string  `json:"bank_name" binding:"required"`
	AccountHolder string  `json:"account_holder" binding:"required"`
	Remark        string  `json:"remark"`
}

// Create 提交提现申请
func (h *WithdrawalHandler) Create(c *gin.Context) {
	// 获取当前登录用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权访问"})
		return
	}

	// 绑定请求参数
	var req CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 创建提现记录
	withdrawal := &model.Withdrawal{
		UserID:        userID.(uint),
		Amount:        req.Amount,
		BankAccount:   req.BankAccount,
		BankName:      req.BankName,
		AccountHolder: req.AccountHolder,
		Remark:        req.Remark,
	}

	if err := h.withdrawalService.CreateWithdrawal(withdrawal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "提现申请提交成功，请等待审核",
		"data": withdrawal,
	})
}

// List 查询提现记录
func (h *WithdrawalHandler) List(c *gin.Context) {
	// 获取当前登录用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权访问"})
		return
	}

	// 查询提现记录
	withdrawals, err := h.withdrawalService.GetUserWithdrawals(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询提现记录失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": withdrawals,
	})
}