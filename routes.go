package main

import (
	"github.com/gin-gonic/gin"
	"github.com/miaomimg3/photo/api/agent"
	"github.com/miaomimg3/photo/api/photo"
	"github.com/miaomimg3/photo/api/user"
)

func setupRoutes(router *gin.Engine) {
	// 用户相关路由
	userHandler := &user.UserHandler{}
	router.POST("/api/register", userHandler.Register)
	router.POST("/api/login", userHandler.Login)

	// 照片相关路由
	photoHandler := &photo.PhotoHandler{}
	router.POST("/api/photo/upload", photoHandler.Upload)
	router.GET("/api/photo/list", photoHandler.List)
	router.GET("/api/photo/:id", photoHandler.Get)

	// 代理相关路由
	agentHandler := &agent.AgentHandler{}
	router.POST("/api/agent/apply", agentHandler.Apply)
	router.POST("/api/agent/withdraw", agentHandler.Withdraw)
	router.GET("/api/agent/withdrawals", agentHandler.ListWithdrawals)

	// 管理员路由
	admin := router.Group("/api/admin")
	// admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admin.GET("/users", userHandler.List)
		admin.GET("/photos", photoHandler.AdminList)
		admin.PUT("/photos/:id/visible", photoHandler.UpdateVisible)
		admin.GET("/withdrawals", agentHandler.AdminListWithdrawals)
		admin.PUT("/withdrawals/:id/status", agentHandler.UpdateWithdrawalStatus)
	}

	router.POST("/api/payment/notify", photoHandler.HandlePaymentNotify)
}