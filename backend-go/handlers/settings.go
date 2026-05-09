package handlers

import (
	"net/http"
	"wholesale/services"

	"github.com/gin-gonic/gin"
)

type DeliveryFeeUpdateRequest struct {
	FreeDistance float64 `json:"free_distance_km"`
	FeePerKM     float64 `json:"fee_per_km_usd"`
}

type DeliveryFeeEstimateRequest struct {
	DistanceKM float64 `json:"distance_km" binding:"min=0"`
}

// GET /api/settings/delivery-fee
func GetDeliveryFee(c *gin.Context) {
	free, fee := services.GetDeliveryFeeSettings()
	c.JSON(http.StatusOK, gin.H{
		"free_distance_km": free,
		"fee_per_km_usd":   fee,
	})
}

// PATCH /api/settings/delivery-fee
func UpdateDeliveryFee(c *gin.Context) {
	var req DeliveryFeeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	services.SaveDeliveryFeeSettings(req.FreeDistance, req.FeePerKM)
	c.JSON(http.StatusOK, gin.H{
		"free_distance_km": req.FreeDistance,
		"fee_per_km_usd":   req.FeePerKM,
	})
}

// POST /api/settings/delivery-fee/estimate
func EstimateDeliveryFee(c *gin.Context) {
	var req DeliveryFeeEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	free, fee := services.GetDeliveryFeeSettings()
	result := services.CalculateDeliveryFee(req.DistanceKM, free, fee)
	c.JSON(http.StatusOK, gin.H{
		"distance_km":      req.DistanceKM,
		"delivery_fee_usd": result,
	})
}

// GET /api/settings/role-chat-ids
func GetRoleChatIDs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"group_chat_id":       services.GetGroupChatID(),
		"delivery_group_link": services.GetDeliveryGroupLink(),
	})
}

// PUT /api/settings/role-chat-ids
func UpdateRoleChatIDs(c *gin.Context) {
	var req struct {
		GroupChatID       string `json:"group_chat_id"`
		DeliveryGroupLink string `json:"delivery_group_link"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	services.SaveGroupChatID(req.GroupChatID)
	services.SaveDeliveryGroupLink(req.DeliveryGroupLink)
	c.JSON(http.StatusOK, gin.H{
		"group_chat_id":       req.GroupChatID,
		"delivery_group_link": req.DeliveryGroupLink,
	})
}

// GET /api/settings/telegram-recent-chats — 从 DB 缓存读取近期 chat 列表
func GetTelegramRecentChats(c *gin.Context) {
	chats := services.GetRecentChats()
	c.JSON(http.StatusOK, gin.H{"chats": chats})
}

// GET /api/settings/contact-info — 公开
func GetContactInfo(c *gin.Context) {
	info := services.GetContactInfo()
	c.JSON(http.StatusOK, gin.H{
		"phone":    info.Phone,
		"telegram": info.Telegram,
		"whatsapp": info.Whatsapp,
		"wechat":   info.Wechat,
	})
}

// PATCH /api/settings/contact-info — 管理员
func UpdateContactInfo(c *gin.Context) {
	var req struct {
		Phone    *string `json:"phone"`
		Telegram *string `json:"telegram"`
		Whatsapp *string `json:"whatsapp"`
		Wechat   *string `json:"wechat"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	// 取当前值，类型为 string
	current := services.GetContactInfo()
	phone := current.Phone
	tg := current.Telegram
	wa := current.Whatsapp
	wc := current.Wechat
	if req.Phone != nil {
		phone = *req.Phone
	}
	if req.Telegram != nil {
		tg = *req.Telegram
	}
	if req.Whatsapp != nil {
		wa = *req.Whatsapp
	}
	if req.Wechat != nil {
		wc = *req.Wechat
	}
	info := services.SaveContactInfo(phone, tg, wa, wc)
	c.JSON(http.StatusOK, gin.H{
		"phone":    info.Phone,
		"telegram": info.Telegram,
		"whatsapp": info.Whatsapp,
		"wechat":   info.Wechat,
	})
}

// GET /api/settings/google-maps  — 管理员
func GetGoogleMapsSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"api_key":       services.GetGoogleMapsKey(),
		"warehouse_lat": services.GetWarehouseLat(),
		"warehouse_lng": services.GetWarehouseLng(),
	})
}

// PATCH /api/settings/google-maps  — 管理员
func UpdateGoogleMapsSettings(c *gin.Context) {
	var req struct {
		APIKey       *string `json:"api_key"`
		WarehouseLat *string `json:"warehouse_lat"`
		WarehouseLng *string `json:"warehouse_lng"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	if req.APIKey != nil {
		services.SaveGoogleMapsKey(*req.APIKey)
	}
	if req.WarehouseLat != nil && req.WarehouseLng != nil {
		services.SaveWarehouseCoords(*req.WarehouseLat, *req.WarehouseLng)
	}
	c.JSON(http.StatusOK, gin.H{
		"api_key":       services.GetGoogleMapsKey(),
		"warehouse_lat": services.GetWarehouseLat(),
		"warehouse_lng": services.GetWarehouseLng(),
	})
}

// POST /api/settings/delivery-fee/estimate-by-address （已登录可用）
func EstimateDeliveryFeeByAddress(c *gin.Context) {
	var req struct {
		Origin      string `json:"origin"` // 可不传，空时自动用仓库坐标
		Destination string `json:"destination" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	// 如果未传起点，使用仓库坐标
	if req.Origin == "" {
		lat := services.GetWarehouseLat()
		lng := services.GetWarehouseLng()
		if lat != "" && lng != "" {
			req.Origin = lat + "," + lng
		}
	}

	apiKey := services.GetGoogleMapsKey()
	distKM, err := services.GetDistanceKM(req.Origin, req.Destination, apiKey)
	if err != nil {
		free, feePerKM := services.GetDeliveryFeeSettings()
		result := services.CalculateDeliveryFee(0, free, feePerKM)
		c.JSON(http.StatusOK, gin.H{
			"distance_km":      0,
			"delivery_fee_usd": result,
			"warning":          "无法获取实际距离，已使用默认计费",
		})
		return
	}

	free, feePerKM := services.GetDeliveryFeeSettings()
	result := services.CalculateDeliveryFee(distKM, free, feePerKM)
	c.JSON(http.StatusOK, gin.H{
		"distance_km":      distKM,
		"delivery_fee_usd": result,
	})
}
