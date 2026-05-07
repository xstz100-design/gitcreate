package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"wholesale/bot"
	"wholesale/config"
	"wholesale/database"
	"wholesale/handlers"
	"wholesale/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置：依次尝试 .env → ../backend/.env
	envFile := ".env"
	for _, candidate := range []string{".env", "../backend/.env"} {
		if _, err := os.Stat(candidate); err == nil {
			envFile = candidate
			break
		}
	}
	config.Load(envFile)

	// 连接数据库
	database.Connect(config.C.DatabaseURL)

	// 启动 Bot 长轮询（非阻塞）
	stopBot := make(chan struct{})
	go bot.StartPolling(config.C.TGBotToken, stopBot)

	// 配置 Gin
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// CORS
	allowedOrigins := config.C.AllowedOrigins
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(allowedOrigins) == 0 {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,Accept")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 全局限流
	r.Use(middleware.RateLimit())

	// 静态文件（上传目录）
	r.Static("/uploads", "./uploads")

	// 健康检查
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "wholesale-go"})
	})

	// ─────────────────── API 路由 ───────────────────
	api := r.Group("/api")

	// ── Auth（公开）
	authPub := api.Group("/auth")
	{
		authPub.POST("/login", handlers.Login)
		authPub.POST("/telegram-auth", handlers.TelegramAuth)
		authPub.POST("/phone-verification/send", handlers.SendPhoneVerification)
		authPub.POST("/phone-verification/verify", handlers.VerifyPhoneCode)
		authPub.POST("/phone-verification/telegram-verify", handlers.TelegramVerifyPhone)
	}

	// ── Auth（需登录）
	authAuth := api.Group("/auth", middleware.Auth())
	{
		authAuth.GET("/me", handlers.GetMe)
		authAuth.PATCH("/me", handlers.UpdateMe)
		authAuth.PATCH("/me/telegram", handlers.UpdateMeTelegram)
		authAuth.POST("/me/telegram/bind-current", handlers.BindCurrentTelegram)
		authAuth.POST("/change-password", handlers.ChangePassword)
		authAuth.POST("/bind-telegram", handlers.BindTelegram)
		authAuth.POST("/submit-review", handlers.SubmitForReview)
	}

	// ── Auth（管理员）
	authAdmin := api.Group("/auth", middleware.Auth(), middleware.RequireAdmin())
	{
		authAdmin.GET("/users", handlers.ListUsers)
		authAdmin.POST("/register", handlers.Register)
		authAdmin.GET("/users/:id", handlers.GetUser)
		authAdmin.PATCH("/users/:id", handlers.UpdateUser)
		authAdmin.DELETE("/users/:id", handlers.DeleteUser)
		authAdmin.POST("/users/:id/approve", handlers.ApproveUser)
		authAdmin.GET("/pending-users", handlers.ListPendingUsers)
		authAdmin.GET("/pending-count", handlers.GetPendingCount)
		authAdmin.GET("/all-registrations", handlers.ListAllRegistrations)
		authAdmin.POST("/users/:id/reset-password", handlers.ResetPassword)
		authAdmin.GET("/dashboard", handlers.GetDashboard)
		authAdmin.GET("/dashboard-metrics", handlers.GetDashboardMetrics)
	}

	// ── Auth（超级管理员）
	authSuper := api.Group("/auth", middleware.Auth(), middleware.RequireSuperAdmin())
	{
		authSuper.PATCH("/users/:id/super-admin", handlers.UpdateSuperAdmin)
		authSuper.POST("/users/:id/super-admin", handlers.SetSuperAdminPost)
	}

	// ── 商品
	productsAuth := api.Group("/products", middleware.Auth())
	{
		productsAuth.GET("", handlers.ListProducts)
		productsAuth.GET("/barcode/:barcode", handlers.GetProductByBarcode)
		productsAuth.GET("/:id", handlers.GetProduct)
	}
	productsAdmin := api.Group("/products", middleware.Auth(), middleware.RequireAdmin())
	{
		productsAdmin.POST("", handlers.CreateProduct)
		productsAdmin.GET("/import/template", handlers.GetImportTemplate)
		productsAdmin.POST("/import", handlers.ImportProducts)
		productsAdmin.PATCH("/:id", handlers.UpdateProduct)
		productsAdmin.DELETE("/:id", handlers.DeleteProduct)
	}

	// ── 订单
	ordersAuth := api.Group("/orders", middleware.Auth())
	{
		ordersAuth.GET("", handlers.ListOrders)
		ordersAuth.POST("", handlers.CreateOrder)
		ordersAuth.GET("/picker/items/:orderId", handlers.GetPickerItems)
		ordersAuth.GET("/:id", handlers.GetOrder)
		ordersAuth.PATCH("/:id", handlers.UpdateOrder)
		ordersAuth.POST("/:id/cancel", handlers.CancelOrder)
		ordersAuth.POST("/:id/pick", handlers.MarkOrderPicked)
	}
	ordersAdmin := api.Group("/orders", middleware.Auth(), middleware.RequireAdmin())
	{
		ordersAdmin.DELETE("/:id", handlers.DeleteOrder)
	}

	// ── 分类
	catPub := api.Group("/categories")
	{
		catPub.GET("", handlers.ListCategories) // 公开，已登录也可访问
	}
	catAuth := api.Group("/categories", middleware.Auth())
	{
		catAuth.GET("/all", handlers.ListAllCategories)
	}
	catAdmin := api.Group("/categories", middleware.Auth(), middleware.RequireAdmin())
	{
		catAdmin.POST("", handlers.CreateCategory)
		catAdmin.PATCH("/:id", handlers.UpdateCategory)
		catAdmin.DELETE("/:id", handlers.DeleteCategory)
	}

	// ── 公告
	annPub := api.Group("/announcements")
	{
		annPub.GET("/public", handlers.ListPublicAnnouncements)
	}
	annAdmin := api.Group("/announcements", middleware.Auth(), middleware.RequireAdmin())
	{
		annAdmin.GET("", handlers.ListAnnouncements)
		annAdmin.POST("", handlers.CreateAnnouncement)
		annAdmin.PATCH("/:id", handlers.UpdateAnnouncement)
		annAdmin.DELETE("/:id", handlers.DeleteAnnouncement)
	}

	// ── 账单
	billingAuth := api.Group("/billing", middleware.Auth())
	{
		billingAuth.GET("", handlers.ListBills)
	}
	billingAdmin := api.Group("/billing", middleware.Auth(), middleware.RequireAdmin())
	{
		billingAdmin.POST("/generate", handlers.GenerateBills)
		billingAdmin.PATCH("/:id", handlers.UpdateBill)
	}

	// ── 设置
	settingsPub := api.Group("/settings")
	{
		settingsPub.GET("/contact-info", handlers.GetContactInfo)
		settingsPub.POST("/delivery-fee/estimate", handlers.EstimateDeliveryFee)
	}
	settingsAuth := api.Group("/settings", middleware.Auth())
	{
		settingsAuth.GET("/delivery-fee", handlers.GetDeliveryFee)
	}
	settingsAdmin := api.Group("/settings", middleware.Auth(), middleware.RequireAdmin())
	{
		settingsAdmin.PATCH("/delivery-fee", handlers.UpdateDeliveryFee)
		settingsAdmin.GET("/role-chat-ids", handlers.GetRoleChatIDs)
		settingsAdmin.PUT("/role-chat-ids", handlers.UpdateRoleChatIDs)
		settingsAdmin.GET("/telegram-recent-chats", handlers.GetTelegramRecentChats)
		settingsAdmin.PATCH("/contact-info", handlers.UpdateContactInfo)
	}

	// ── 图片上传
	api.POST("/upload/image", middleware.Auth(), handlers.UploadImage)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	fmt.Printf("🚀 wholesale-go 启动中，监听 %s\n", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
