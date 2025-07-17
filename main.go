package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/miaoming3/photo/internal/config"
	"github.com/miaoming3/photo/internal/handler"
	"github.com/miaoming3/photo/internal/middleware"
	"github.com/miaoming3/photo/internal/model"
	"github.com/miaoming3/photo/internal/service"
)

func main() {
	// 初始化配置
	if err := config.Init("configs/app.yaml"); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 初始化数据库
	db, err := initDB()
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// 初始化服务
	fileService := service.NewFileService(db)
	orderService := service.NewOrderService(db)
	userService := service.NewUserService(db)
	paymentGateway := &service.MockPaymentGateway{}
	paymentService := service.NewPaymentService(db, paymentGateway)
	withdrawalService := service.NewWithdrawalService(db)
	notificationService := service.NewNotificationService(db)

	// 初始化处理器
	fileHandler := handler.NewFileHandler(fileService, orderService)
	orderHandler := handler.NewOrderHandler(orderService, paymentService)
	userHandler := handler.NewUserHandler(userService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	withdrawalHandler := handler.NewWithdrawalHandler(withdrawalService)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// 设置路由
	router := gin.Default()
	router.Static("/uploads", config.Cfg.Storage.RootPath)
	router.Static("/web", "./web")
	api := router.Group("/api")
	{
		// 用户相关路由（无需认证）
		api.POST("/register", userHandler.Register)
		api.POST("/login", userHandler.Login)

		// 需要认证的路由组
		auth := api.Group("/")
		auth.Use(middleware.AuthMiddleware())
		{
			auth.POST("/upload", fileHandler.Upload)
			auth.GET("/files", fileHandler.List)
			auth.DELETE("/files/:id", fileHandler.Delete)
			auth.POST("/orders", orderHandler.Create)
			auth.GET("/orders/status", orderHandler.GetStatus)
			auth.GET("/orders", orderHandler.List)
			auth.POST("/withdrawals", withdrawalHandler.Create)
			auth.GET("/withdrawals", withdrawalHandler.List)

			// 添加通知路由

			auth.GET("/notifications", notificationHandler.List)
			auth.GET("/notifications/unread-count", notificationHandler.UnreadCount)
			auth.PUT("/notifications/:id/read", notificationHandler.MarkAsRead)
		}
		// 支付回调路由（不需要认证）
		router.POST("/payments/callback", paymentHandler.HandleCallback)
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:           ":" + config.Cfg.Server.Port,
		Handler:        router,
		ReadTimeout:    config.Cfg.Server.ReadTimeout,
		WriteTimeout:   config.Cfg.Server.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// 启动服务器
	go func() {
		log.Printf("服务器启动在端口 %s\n", config.Cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器关闭失败:", err)
	}

	log.Println("服务器已关闭")
}

// 初始化数据库连接
func initDB() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(config.Cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 迁移数据表
	if err := db.Set("gorm:table_options", "ENGINE=InnoDB").
		AutoMigrate(&model.File{}, &model.User{}, &model.Order{}, &model.Withdrawal{}, &model.Notification{}); err != nil {
		return nil, err
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(config.Cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.Cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.Cfg.Database.ConnMaxLifetime)

	return db, nil
}
