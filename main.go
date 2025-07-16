package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/miaomimg3/photo/config"
	"github.com/miaomimg3/photo/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库连接
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	models.DB = db

	// 自动迁移数据表
	// 自动迁移数据表
	DB.AutoMigrate(&models.User{}, &models.Photo{}, &models.AgentWithdrawal{}, &models.Order{})

	// 设置路由
	router := gin.Default()
	setupRoutes(router)

	// 启动服务器
	fmt.Printf("服务器启动在端口 %s\n", cfg.Server.Port)
	log.Fatal(router.Run(":" + cfg.Server.Port))
}
