package agent

import (
	"github.com/gin-gonic/gin"
	"github.com/miaomimg3/photo/models"
	"net/http"
	"time"
)

type AgentHandler struct{}

// Apply 申请成为代理
func (h *AgentHandler) Apply(c *gin.Context) {
	var req struct {
		UserID   uint   `json:"user_id" binding:"required"`
		BankInfo string `json:"bank_info" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新用户为代理
	if err := models.DB.Model(&models.User{}).Where("id = ?", req.UserID).Update("is_agent", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "申请代理失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "代理申请成功"})
}

// Withdraw 代理提现
func (h *AgentHandler) Withdraw(c *gin.Context) {
	var req struct {
		AgentID  uint    `json:"agent_id" binding:"required"`
		Amount   float64 `json:"amount" binding:"required,min=1"`
		BankInfo string  `json:"bank_info" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否为代理
	var user models.User
	if err := models.DB.Where("id = ? AND is_agent = ?", req.AgentID, true).First(&user).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "非代理用户无法提现"})
		return
	}

	// 检查余额是否充足
	if user.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "余额不足"})
		return
	}

	// 创建提现记录
	withdrawal := models.AgentWithdrawal{
		AgentID:   req.AgentID,
		Amount:    req.Amount,
		Status:    "pending",
		BankInfo:  req.BankInfo,
		CreatedAt: time.Now().Unix(),
	}

	if err := models.DB.Create(&withdrawal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建提现记录失败"})
		return
	}

	// 扣减余额
	models.DB.Model(&user).Update("balance", user.Balance - req.Amount)

	c.JSON(http.StatusOK, gin.H{"message": "提现申请已提交", "withdrawal_id": withdrawal.ID})
}

// ListWithdrawals 获取代理提现记录
func (h *AgentHandler) ListWithdrawals(c *gin.Context) {
	agentID := c.Query("agent_id")
	var withdrawals []models.AgentWithdrawal
	if err := models.DB.Where("agent_id = ?", agentID).Find(&withdrawals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取提现记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"withdrawals": withdrawals})
}

// AdminListWithdrawals 管理员获取所有提现记录
func (h *AgentHandler) AdminListWithdrawals(c *gin.Context) {
	var withdrawals []models.AgentWithdrawal
	if err := models.DB.Find(&withdrawals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取提现记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"withdrawals": withdrawals})
}

// UpdateWithdrawalStatus 管理员更新提现状态
func (h *AgentHandler) UpdateWithdrawalStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required,oneof=pending approved rejected"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := models.DB.Model(&models.AgentWithdrawal{}).Where("id = ?", id).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新提现状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "提现状态已更新"})
}