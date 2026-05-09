package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"wholesale/database"
	"wholesale/models"
)

// ─────────────────── 活跃用户内存去重 ───────────────────

var (
	auMu   sync.Mutex
	auDate string         // 当前记录的日期（柬埔寨时区 yyyy-mm-dd）
	auSet  map[int64]bool // 今日已记录的用户 ID
)

const (
	KeyDeliveryFreeDistance = "delivery.free_distance_km"
	KeyDeliveryFeePerKM     = "delivery.fee_per_extra_km_usd"
	KeyGroupChatID          = "telegram.group_chat_id"
	KeyContactPhone         = "contact.phone"
	KeyContactTelegram      = "contact.telegram"
	KeyContactWhatsapp      = "contact.whatsapp"
	KeyContactWechat        = "contact.wechat"
	KeyGoogleMapsKey        = "delivery.google_maps_key"
	KeyWarehouseAddress     = "delivery.warehouse_address"
	KeyWarehouseLat         = "delivery.warehouse_lat"
	KeyWarehouseLng         = "delivery.warehouse_lng"
	KeyDeliveryGroupLink    = "telegram.delivery_group_link"
	KeyRecentChats          = "telegram.recent_chats"

	DefaultFreeDistanceKM = 3.0
	DefaultFeePerKM       = 0.5
)

// RecentChat 记录 bot 收到消息的 Telegram chat（用于管理员选择群）
type RecentChat struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// GetRecentChats 从 DB 读取缓存的近期 chat 列表
func GetRecentChats() []RecentChat {
	raw := GetSetting(KeyRecentChats)
	if raw == "" {
		return []RecentChat{}
	}
	var chats []RecentChat
	if err := json.Unmarshal([]byte(raw), &chats); err != nil {
		return []RecentChat{}
	}
	return chats
}

// UpsertRecentChat 更新或新增一条近期 chat 记录（最多保留 50 条，群/频道优先排前）
func UpsertRecentChat(chatID int64, title, chatType string) {
	idStr := fmt.Sprintf("%d", chatID)
	chats := GetRecentChats()
	// 已存在 → 更新并移到最前
	for i, c := range chats {
		if c.ID == idStr {
			chats[i].Title = title
			chats[i].Type = chatType
			// 移到最前
			chats = append([]RecentChat{chats[i]}, append(chats[:i], chats[i+1:]...)...)
			b, _ := json.Marshal(chats)
			UpsertSetting(KeyRecentChats, string(b))
			return
		}
	}
	// 新增到最前
	chats = append([]RecentChat{{ID: idStr, Title: title, Type: chatType}}, chats...)
	if len(chats) > 50 {
		chats = chats[:50]
	}
	b, _ := json.Marshal(chats)
	UpsertSetting(KeyRecentChats, string(b))
}

func GetSetting(key string) string {
	var s models.SystemSetting
	if err := database.DB.Where("key = ?", key).First(&s).Error; err != nil {
		return ""
	}
	return s.Value
}

func UpsertSetting(key, value string) {
	var s models.SystemSetting
	if database.DB.Where("key = ?", key).First(&s).Error == nil {
		database.DB.Model(&s).Update("value", value)
		database.DB.Model(&s).Update("updated_at", models.NowCambodia())
	} else {
		database.DB.Create(&models.SystemSetting{Key: key, Value: value, UpdatedAt: models.NowCambodia()})
	}
}

// ─────────────────── 配送费 ───────────────────

func GetDeliveryFeeSettings() (freeKM, feePerKM float64) {
	freeKM = parseFloatOrDefault(GetSetting(KeyDeliveryFreeDistance), DefaultFreeDistanceKM)
	feePerKM = parseFloatOrDefault(GetSetting(KeyDeliveryFeePerKM), DefaultFeePerKM)
	if freeKM < 0 {
		freeKM = 0
	}
	if feePerKM < 0 {
		feePerKM = 0
	}
	return
}

func SaveDeliveryFeeSettings(freeKM, feePerKM float64) {
	if freeKM < 0 {
		freeKM = 0
	}
	if feePerKM < 0 {
		feePerKM = 0
	}
	UpsertSetting(KeyDeliveryFreeDistance, strconv.FormatFloat(freeKM, 'f', 2, 64))
	UpsertSetting(KeyDeliveryFeePerKM, strconv.FormatFloat(feePerKM, 'f', 2, 64))
}

func CalculateDeliveryFee(distanceKM, freeKM, feePerKM float64) float64 {
	if distanceKM < 0 {
		distanceKM = 0
	}
	if distanceKM <= freeKM {
		return 0
	}
	extra := distanceKM - freeKM
	// 向上取整到整公里
	extraUnits := int(extra)
	if float64(extraUnits) < extra {
		extraUnits++
	}
	fee := float64(extraUnits) * feePerKM
	// 保留2位小数
	fee, _ = strconv.ParseFloat(strconv.FormatFloat(fee, 'f', 2, 64), 64)
	return fee
}

// ─────────────────── Telegram Chat IDs ───────────────────

func GetGroupChatID() string {
	return GetSetting(KeyGroupChatID)
}

func SaveGroupChatID(id string) {
	UpsertSetting(KeyGroupChatID, id)
}

// ─────────────────── 联系方式 ───────────────────

type ContactInfo struct {
	Phone    string `json:"phone"`
	Telegram string `json:"telegram"`
	Whatsapp string `json:"whatsapp"`
	Wechat   string `json:"wechat"`
}

func GetContactInfo() ContactInfo {
	return ContactInfo{
		Phone:    GetSetting(KeyContactPhone),
		Telegram: GetSetting(KeyContactTelegram),
		Whatsapp: GetSetting(KeyContactWhatsapp),
		Wechat:   GetSetting(KeyContactWechat),
	}
}

func SaveContactInfo(phone, telegram, whatsapp, wechat string) ContactInfo {
	UpsertSetting(KeyContactPhone, phone)
	UpsertSetting(KeyContactTelegram, telegram)
	UpsertSetting(KeyContactWhatsapp, whatsapp)
	UpsertSetting(KeyContactWechat, wechat)
	return GetContactInfo()
}

// ─────────────────── Google Maps ───────────────────

func GetGoogleMapsKey() string {
	return GetSetting(KeyGoogleMapsKey)
}

func SaveGoogleMapsKey(key string) {
	UpsertSetting(KeyGoogleMapsKey, key)
}

func GetWarehouseAddress() string {
	return GetSetting(KeyWarehouseAddress)
}

func SaveWarehouseAddress(addr string) {
	UpsertSetting(KeyWarehouseAddress, addr)
}

func GetWarehouseLat() string {
	return GetSetting(KeyWarehouseLat)
}

func GetWarehouseLng() string {
	return GetSetting(KeyWarehouseLng)
}

func SaveWarehouseCoords(lat, lng string) {
	UpsertSetting(KeyWarehouseLat, lat)
	UpsertSetting(KeyWarehouseLng, lng)
}

func GetDeliveryGroupLink() string {
	return GetSetting(KeyDeliveryGroupLink)
}

func SaveDeliveryGroupLink(link string) {
	UpsertSetting(KeyDeliveryGroupLink, link)
}

// ─────────────────── 统计 ───────────────────

// TrackPageView 记录页面访问（异步，不阻塞请求）
func TrackPageView() {
	go func() {
		today := models.NowCambodia().Format("2006-01-02")
		database.DB.Exec(`
			INSERT INTO daily_metrics (date, page_views, active_users, updated_at)
			VALUES (?, 1, 0, ?)
			ON CONFLICT(date) DO UPDATE SET page_views = page_views + 1, updated_at = ?
		`, today, models.NowCambodia(), models.NowCambodia())
	}()
}

// TrackActiveUser 记录活跃用户（每用户每天只计一次，使用内存去重避免污染 DB）
func TrackActiveUser(userID int64) {
	today := models.NowCambodia().Format("2006-01-02")

	auMu.Lock()
	if auDate != today {
		// 新的一天，重置集合
		auDate = today
		auSet = make(map[int64]bool)
	}
	if auSet[userID] {
		auMu.Unlock()
		return // 今天已记录过，跳过
	}
	auSet[userID] = true
	auMu.Unlock()

	// 今日首次活跃，写入 DB（已在 goroutine 调用方异步执行）
	database.DB.Exec(`
		INSERT INTO daily_metrics (date, page_views, active_users, updated_at)
		VALUES (?, 0, 1, ?)
		ON CONFLICT(date) DO UPDATE SET active_users = active_users + 1, updated_at = ?
	`, today, models.NowCambodia(), models.NowCambodia())
}

// GetMetrics 获取最近 N 天的统计
func GetMetrics(days int) []models.DailyMetric {
	if days <= 0 {
		days = 7
	}
	var metrics []models.DailyMetric
	database.DB.Order("date DESC").Limit(days).Find(&metrics)
	return metrics
}

// ─────────────────── helpers ───────────────────

func parseFloatOrDefault(s string, def float64) float64 {
	if s == "" {
		return def
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return f
}
