package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"wholesale/config"
	"wholesale/database"
	"wholesale/middleware"
	"wholesale/models"
	"wholesale/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var phoneRegexp = regexp.MustCompile(`^[+\d][\d\s\-]{5,19}$`)

// ─────────────────── 请求 / 响应结构 ───────────────────

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TelegramAuthRequest struct {
	InitData string `json:"init_data" binding:"required"`
}

type TelegramBindRequest struct {
	InitData string `json:"init_data" binding:"required"`
}

type UserCreateRequest struct {
	FullName         string          `json:"full_name" binding:"required"`
	Role             models.UserRole `json:"role"`
	Phone            *string         `json:"phone"`
	Address          *string         `json:"address"`
	LocationURL      *string         `json:"location_url"`
	CreditLimit      float64         `json:"credit_limit"`
	BillingCycleDays *int            `json:"billing_cycle_days"`
	AllowCredit      bool            `json:"allow_credit"`
}

type UserUpdateRequest struct {
	FullName         *string  `json:"full_name"`
	Phone            *string  `json:"phone"`
	Address          *string  `json:"address"`
	LocationURL      *string  `json:"location_url"`
	StorePhoto       *string  `json:"store_photo"`
	CreditLimit      *float64 `json:"credit_limit"`
	BillingCycleDays *int     `json:"billing_cycle_days"`
	AllowCredit      *bool    `json:"allow_credit"`
	NotifyEnabled    *bool    `json:"notify_enabled"`
	IsActive         *bool    `json:"is_active"`
}

type PasswordChangeRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ApprovalRequest struct {
	Approved       bool    `json:"approved"`
	RejectedReason *string `json:"rejected_reason"`
}

type SuperAdminUpdateRequest struct {
	IsSuperAdmin bool `json:"is_super_admin"`
}

type UserResponse struct {
	ID                 int64                 `json:"id"`
	Username           string                `json:"username"`
	FullName           string                `json:"full_name"`
	Role               models.UserRole       `json:"role"`
	Phone              *string               `json:"phone"`
	Address            *string               `json:"address"`
	CreditLimit        float64               `json:"credit_limit"`
	BillingCycleDays   *int                  `json:"billing_cycle_days"`
	AllowCredit        bool                  `json:"allow_credit"`
	LocationURL        *string               `json:"location_url"`
	StorePhoto         *string               `json:"store_photo"`
	TelegramID         *int64                `json:"telegram_id"`
	NotifyEnabled      bool                  `json:"notify_enabled"`
	MustChangePassword bool                  `json:"must_change_password"`
	IsActive           bool                  `json:"is_active"`
	ApprovalStatus     models.ApprovalStatus `json:"approval_status"`
	RejectedReason     *string               `json:"rejected_reason"`
	ProfileCompleted   bool                  `json:"profile_completed"`
	IsSuperAdmin       bool                  `json:"is_super_admin"`
}

type TokenResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	User        UserResponse `json:"user"`
}

type UserCreateResponse struct {
	User              UserResponse `json:"user"`
	TemporaryPassword string       `json:"temporary_password"`
}

// buildUserResponse 从 User 模型构建响应 DTO
func buildUserResponse(u *models.User) UserResponse {
	profileCompleted := u.FullName != "" &&
		!strings.HasPrefix(u.FullName, "TG_") &&
		u.Phone != nil && *u.Phone != "" &&
		u.Address != nil && *u.Address != ""

	return UserResponse{
		ID:                 u.ID,
		Username:           u.Username,
		FullName:           u.FullName,
		Role:               u.Role,
		Phone:              u.Phone,
		Address:            u.Address,
		CreditLimit:        u.CreditLimit,
		BillingCycleDays:   u.BillingCycleDays,
		AllowCredit:        u.AllowCredit,
		LocationURL:        u.LocationURL,
		StorePhoto:         u.StorePhoto,
		TelegramID:         u.TelegramID,
		NotifyEnabled:      u.NotifyEnabled,
		MustChangePassword: u.MustChangePassword,
		IsActive:           u.IsActive,
		ApprovalStatus:     u.ApprovalStatus,
		RejectedReason:     u.RejectedReason,
		ProfileCompleted:   profileCompleted,
		IsSuperAdmin:       u.IsSuperAdmin,
	}
}

// ─────────────────── 账号生成 ───────────────────

func generateAccountNumber(role models.UserRole) (string, error) {
	var start, end int
	if role == models.RoleAdmin {
		start, end = 100001, 200000
	} else {
		start, end = 200001, 299999
	}

	var maxNum *int
	row := database.DB.Raw(`
		SELECT MAX(CAST(username AS INTEGER)) FROM users
		WHERE CAST(username AS INTEGER) >= ? AND CAST(username AS INTEGER) <= ?
	`, start, end).Scan(&maxNum)
	if row.Error != nil {
		return "", row.Error
	}

	next := start
	if maxNum != nil {
		next = *maxNum + 1
	}
	if next > end {
		return "", fmt.Errorf("账号号段已用尽")
	}
	return strconv.Itoa(next), nil
}

// ─────────────────── 登录 ───────────────────

// POST /api/auth/login (form-urlencoded: username + password)
func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		// 也支持 JSON body
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "请提供用户名和密码"})
			return
		}
		username, password = req.Username, req.Password
	}

	ip := c.ClientIP()
	if err := middleware.LoginGuard.Check(ip, username); err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"detail": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		middleware.LoginGuard.RecordFailure(ip, username)
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "用户名或密码错误"})
		return
	}

	if !utils.VerifyPassword(password, user.HashedPassword) {
		middleware.LoginGuard.RecordFailure(ip, username)
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "用户名或密码错误"})
		return
	}

	if user.Role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"detail": "商户请通过 Telegram Mini App 登录"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "账号已被禁用"})
		return
	}

	middleware.LoginGuard.Reset(ip, username)
	token, err := utils.CreateAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "生成 Token 失败"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
		User:        buildUserResponse(&user),
	})
}

// ─────────────────── Telegram 免登录 ───────────────────

// POST /api/auth/telegram-auth
func TelegramAuth(c *gin.Context) {
	var req TelegramAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	tgUser, err := validateTelegramInitData(req.InitData)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Telegram 验证失败: " + err.Error()})
		return
	}

	tgID := tgUser["id"].(int64)

	var user models.User
	err = database.DB.Where("telegram_id = ?", tgID).First(&user).Error
	if err != nil {
		// 自动创建新用户
		firstName, _ := tgUser["first_name"].(string)
		lastName, _ := tgUser["last_name"].(string)
		fullName := strings.TrimSpace(firstName + " " + lastName)
		if fullName == "" {
			fullName = fmt.Sprintf("TG_%d", tgID)
		}

		username := fmt.Sprintf("tg_%d", tgID)
		hashed, _ := utils.HashPassword(fmt.Sprintf("tg_auto_%d", tgID))

		user = models.User{
			Username:           username,
			HashedPassword:     hashed,
			FullName:           fullName,
			Role:               models.RoleMerchant,
			TelegramID:         &tgID,
			ApprovalStatus:     models.ApprovalPending,
			MustChangePassword: false,
			IsActive:           true,
			NotifyEnabled:      true,
			CreatedAt:          models.NowCambodia(),
		}
		if err2 := database.DB.Create(&user).Error; err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "创建用户失败"})
			return
		}
	}

	if !user.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "账号已被禁用"})
		return
	}

	token, err := utils.CreateAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "生成 Token 失败"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
		User:        buildUserResponse(&user),
	})
}

// ─────────────────── 获取/更新当前用户 ───────────────────

// GET /api/auth/me
func GetMe(c *gin.Context) {
	user := middleware.CurrentUser(c)
	c.JSON(http.StatusOK, buildUserResponse(user))
}

// PATCH /api/auth/me
func UpdateMe(c *gin.Context) {
	user := middleware.CurrentUser(c)

	var req UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.FullName != nil {
		updates["full_name"] = *req.FullName
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.LocationURL != nil {
		updates["location_url"] = *req.LocationURL
	}
	if req.StorePhoto != nil {
		updates["store_photo"] = *req.StorePhoto
	}
	if req.NotifyEnabled != nil {
		updates["notify_enabled"] = *req.NotifyEnabled
	}

	if len(updates) > 0 {
		database.DB.Model(user).Updates(updates)
		database.DB.First(user, user.ID)
	}

	c.JSON(http.StatusOK, buildUserResponse(user))
}

// POST /api/auth/change-password
func ChangePassword(c *gin.Context) {
	user := middleware.CurrentUser(c)

	var req PasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	if !utils.VerifyPassword(req.OldPassword, user.HashedPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "旧密码错误"})
		return
	}

	hashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "密码处理失败"})
		return
	}

	database.DB.Model(user).Updates(map[string]interface{}{
		"hashed_password":      hashed,
		"must_change_password": false,
	})
	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// ─────────────────── 管理员: 用户列表 ───────────────────

// GET /api/auth/users
func ListUsers(c *gin.Context) {
	var users []models.User
	q := database.DB.Where("is_active = ?", true)
	if role := c.Query("role"); role != "" {
		q = q.Where("role = ?", role)
	}
	q.Order("created_at DESC").Find(&users)

	result := make([]UserResponse, len(users))
	for i, u := range users {
		u := u
		result[i] = buildUserResponse(&u)
	}
	c.JSON(http.StatusOK, result)
}

// POST /api/auth/register - 管理员创建用户
func Register(c *gin.Context) {
	admin := middleware.CurrentUser(c)

	var req UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	if req.Role == "" {
		req.Role = models.RoleMerchant
	}

	// 只有超级管理员能创建管理员账号
	if req.Role == models.RoleAdmin && !admin.IsSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"detail": "仅超级管理员可创建管理员账号"})
		return
	}

	var username string
	if req.Role == models.RoleMerchant {
		if req.Phone == nil || *req.Phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "商户必须填写手机号"})
			return
		}
		phone := strings.TrimSpace(*req.Phone)
		if !phoneRegexp.MatchString(phone) {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "手机号格式不正确"})
			return
		}
		var existing models.User
		if database.DB.Where("username = ?", phone).First(&existing).Error == nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "该手机号已被注册"})
			return
		}
		username = phone
	} else {
		var err error
		username, err = generateAccountNumber(req.Role)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}
	}

	tmpPwd := utils.GenerateTemporaryPassword(10)
	hashed, _ := utils.HashPassword(tmpPwd)

	user := models.User{
		Username:           username,
		HashedPassword:     hashed,
		FullName:           req.FullName,
		Role:               req.Role,
		Phone:              req.Phone,
		Address:            req.Address,
		LocationURL:        req.LocationURL,
		CreditLimit:        req.CreditLimit,
		BillingCycleDays:   req.BillingCycleDays,
		AllowCredit:        req.AllowCredit,
		ApprovalStatus:     models.ApprovalApproved,
		MustChangePassword: true,
		IsActive:           true,
		NotifyEnabled:      true,
		CreatedAt:          models.NowCambodia(),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "创建用户失败"})
		return
	}

	c.JSON(http.StatusCreated, UserCreateResponse{
		User:              buildUserResponse(&user),
		TemporaryPassword: tmpPwd,
	})
}

// GET /api/auth/users/:id
func GetUser(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, buildUserResponse(&user))
}

// PATCH /api/auth/users/:id - 管理员更新用户
func UpdateUser(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "用户不存在"})
		return
	}

	var req UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.FullName != nil {
		updates["full_name"] = *req.FullName
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.LocationURL != nil {
		updates["location_url"] = *req.LocationURL
	}
	if req.StorePhoto != nil {
		updates["store_photo"] = *req.StorePhoto
	}
	if req.CreditLimit != nil {
		updates["credit_limit"] = *req.CreditLimit
	}
	if req.BillingCycleDays != nil {
		updates["billing_cycle_days"] = *req.BillingCycleDays
	}
	if req.AllowCredit != nil {
		updates["allow_credit"] = *req.AllowCredit
	}
	if req.NotifyEnabled != nil {
		updates["notify_enabled"] = *req.NotifyEnabled
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		database.DB.Model(&user).Updates(updates)
		database.DB.First(&user, id)
	}
	c.JSON(http.StatusOK, buildUserResponse(&user))
}

// POST /api/auth/users/:id/approve - 审核用户
func ApproveUser(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "用户不存在"})
		return
	}

	var req ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	now := models.NowCambodia()
	updates := map[string]interface{}{}
	if req.Approved {
		updates["approval_status"] = models.ApprovalApproved
		updates["approved_at"] = now
		updates["rejected_reason"] = nil
	} else {
		updates["approval_status"] = models.ApprovalRejected
		updates["rejected_reason"] = req.RejectedReason
	}
	database.DB.Model(&user).Updates(updates)
	database.DB.First(&user, id)
	c.JSON(http.StatusOK, buildUserResponse(&user))
}

// GET /api/auth/pending-users - 待审核用户列表
func ListPendingUsers(c *gin.Context) {
	var users []models.User
	database.DB.Where("approval_status = ?", models.ApprovalPending).Order("created_at ASC").Find(&users)

	type PendingUserResponse struct {
		ID             int64                 `json:"id"`
		Username       string                `json:"username"`
		FullName       string                `json:"full_name"`
		Phone          *string               `json:"phone"`
		Address        *string               `json:"address"`
		ApprovalStatus models.ApprovalStatus `json:"approval_status"`
		RejectedReason *string               `json:"rejected_reason"`
		CreatedAt      string                `json:"created_at"`
	}

	result := make([]PendingUserResponse, len(users))
	for i, u := range users {
		result[i] = PendingUserResponse{
			ID:             u.ID,
			Username:       u.Username,
			FullName:       u.FullName,
			Phone:          u.Phone,
			Address:        u.Address,
			ApprovalStatus: u.ApprovalStatus,
			RejectedReason: u.RejectedReason,
			CreatedAt:      u.CreatedAt.Format(time.RFC3339),
		}
	}
	c.JSON(http.StatusOK, result)
}

// PATCH /api/auth/users/:id/super-admin - 设置超级管理员
func UpdateSuperAdmin(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "用户不存在"})
		return
	}

	var req SuperAdminUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	// 至少保留一个超级管理员
	if !req.IsSuperAdmin && user.IsSuperAdmin {
		var count int64
		database.DB.Model(&models.User{}).Where("role = ? AND is_super_admin = ?", models.RoleAdmin, true).Count(&count)
		if count <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "至少保留一个超级管理员"})
			return
		}
	}

	database.DB.Model(&user).Update("is_super_admin", req.IsSuperAdmin)
	database.DB.First(&user, id)
	c.JSON(http.StatusOK, buildUserResponse(&user))
}

// POST /api/auth/bind-telegram - 绑定当前 Telegram 到已登录账号
func BindTelegram(c *gin.Context) {
	user := middleware.CurrentUser(c)

	var req TelegramBindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	tgUser, err := validateTelegramInitData(req.InitData)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Telegram 验证失败: " + err.Error()})
		return
	}

	tgID := tgUser["id"].(int64)

	// 检查是否已被其他用户绑定
	var existing models.User
	if database.DB.Where("telegram_id = ? AND id != ?", tgID, user.ID).First(&existing).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "该 Telegram 账号已绑定其他用户"})
		return
	}

	database.DB.Model(user).Update("telegram_id", tgID)
	database.DB.First(user, user.ID)
	c.JSON(http.StatusOK, buildUserResponse(user))
}

// ─────────────────── Dashboard 统计 ───────────────────

// GET /api/auth/dashboard
func GetDashboard(c *gin.Context) {
	var totalOrders, pendingOrders, deliveredOrders int64
	var totalRevenue float64
	var totalMerchants, pendingApprovals int64

	database.DB.Model(&models.Order{}).Where("is_deleted = ?", false).Count(&totalOrders)
	database.DB.Model(&models.Order{}).Where("delivery_status = ? AND is_deleted = ?", models.DeliveryPending, false).Count(&pendingOrders)
	database.DB.Model(&models.Order{}).Where("delivery_status = ? AND is_deleted = ?", models.DeliveryDelivered, false).Count(&deliveredOrders)
	database.DB.Model(&models.Order{}).Where("is_deleted = ?", false).Select("COALESCE(SUM(total_usd), 0)").Scan(&totalRevenue)
	database.DB.Model(&models.User{}).Where("role = ?", models.RoleMerchant).Count(&totalMerchants)
	database.DB.Model(&models.User{}).Where("approval_status = ?", models.ApprovalPending).Count(&pendingApprovals)

	var lowStockCount int64
	database.DB.Model(&models.Product{}).Where("is_deleted = ? AND stock <= stock_warning", false).Count(&lowStockCount)

	c.JSON(http.StatusOK, gin.H{
		"total_orders":      totalOrders,
		"pending_orders":    pendingOrders,
		"delivered_orders":  deliveredOrders,
		"total_revenue_usd": totalRevenue,
		"total_merchants":   totalMerchants,
		"pending_approvals": pendingApprovals,
		"low_stock_count":   lowStockCount,
	})
}

// ─────────────────── Telegram initData 验证 ───────────────────

func validateTelegramInitData(initData string) (map[string]interface{}, error) {
	token := config.C.TGBotToken
	if token == "" {
		return nil, fmt.Errorf("TG_BOT_TOKEN 未配置")
	}

	parsed, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("解析 initData 失败")
	}

	receivedHash := parsed.Get("hash")
	if receivedHash == "" {
		return nil, fmt.Errorf("缺少 hash 字段")
	}

	// 构建 data-check-string
	var pairs []string
	for k, v := range parsed {
		if k != "hash" {
			pairs = append(pairs, k+"="+v[0])
		}
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// 计算 HMAC
	secretKey := hmacSHA256([]byte("WebAppData"), []byte(token))
	computedHash := hex.EncodeToString(hmacSHA256(secretKey, []byte(dataCheckString)))

	if !hmac.Equal([]byte(computedHash), []byte(receivedHash)) {
		return nil, fmt.Errorf("签名验证失败")
	}

	// 验证过期 (24h)
	authDateStr := parsed.Get("auth_date")
	if authDateStr != "" {
		authDate, _ := strconv.ParseInt(authDateStr, 10, 64)
		if time.Now().Unix()-authDate > 86400 {
			return nil, fmt.Errorf("initData 已过期")
		}
	}

	// 解析 user JSON
	userStr := parsed.Get("user")
	if userStr == "" {
		return nil, fmt.Errorf("缺少 user 字段")
	}

	var rawUser map[string]interface{}
	if err := json.Unmarshal([]byte(userStr), &rawUser); err != nil {
		return nil, fmt.Errorf("解析 user 失败")
	}

	// 统一 id 为 int64
	if idFloat, ok := rawUser["id"].(float64); ok {
		rawUser["id"] = int64(idFloat)
	}

	return rawUser, nil
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// ─────────────────── 重置密码（管理员） ───────────────────

// POST /api/auth/users/:id/reset-password
func ResetPassword(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "用户不存在"})
		return
	}

	tmpPwd := utils.GenerateTemporaryPassword(10)
	hashed, _ := utils.HashPassword(tmpPwd)
	database.DB.Model(&user).Updates(map[string]interface{}{
		"hashed_password":      hashed,
		"must_change_password": true,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":            "密码已重置",
		"temporary_password": tmpPwd,
	})
}

// 辅助: 判断 gin.Context 是否有 admin 权限
func isAdmin(c *gin.Context) bool {
	u := middleware.CurrentUser(c)
	return u != nil && u.Role == models.RoleAdmin
}

func isSuperAdmin(c *gin.Context) bool {
	u := middleware.CurrentUser(c)
	return u != nil && u.IsSuperAdmin
}

// 确保 gorm 导入被使用
var _ = gorm.ErrRecordNotFound

// ─────────────────── Dashboard Metrics ───────────────────

// GET /api/auth/dashboard-metrics
func GetDashboardMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"active_users": 0,
		"page_views":   0,
	})
}

// ─────────────────── 全部注册用户（含各审核状态） ───────────────────

// GET /api/auth/all-registrations?status=pending|approved|rejected
func ListAllRegistrations(c *gin.Context) {
	var users []models.User
	q := database.DB.Order("created_at DESC")
	if status := c.Query("status"); status != "" {
		q = q.Where("approval_status = ?", status)
	}
	q.Find(&users)

	result := make([]UserResponse, len(users))
	for i, u := range users {
		u := u
		result[i] = buildUserResponse(&u)
	}
	c.JSON(http.StatusOK, result)
}

// ─────────────────── 待审核用户数 ───────────────────

// GET /api/auth/pending-count
func GetPendingCount(c *gin.Context) {
	var count int64
	database.DB.Model(&models.User{}).Where("approval_status = ?", models.ApprovalPending).Count(&count)
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// ─────────────────── 提交/重新提交审核 ───────────────────

// POST /api/auth/submit-review
func SubmitForReview(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user.ApprovalStatus == models.ApprovalApproved {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "账号已通过审核"})
		return
	}
	database.DB.Model(user).Updates(map[string]interface{}{
		"approval_status": models.ApprovalPending,
		"rejected_reason": nil,
	})
	database.DB.First(user, user.ID)
	c.JSON(http.StatusOK, buildUserResponse(user))
}

// ─────────────────── 更新当前用户 Telegram ───────────────────

type TelegramUpdateRequest struct {
	TelegramID *int64 `json:"telegram_id"`
}

// PATCH /api/auth/me/telegram
func UpdateMeTelegram(c *gin.Context) {
	user := middleware.CurrentUser(c)
	var req TelegramUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	if req.TelegramID != nil {
		// 检查是否被其他用户占用
		var existing models.User
		if database.DB.Where("telegram_id = ? AND id != ?", *req.TelegramID, user.ID).First(&existing).Error == nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "该 Telegram 账号已绑定其他用户"})
			return
		}
		database.DB.Model(user).Update("telegram_id", *req.TelegramID)
	} else {
		database.DB.Model(user).Update("telegram_id", nil)
	}
	database.DB.First(user, user.ID)
	c.JSON(http.StatusOK, buildUserResponse(user))
}

// POST /api/auth/me/telegram/bind-current — 与 BindTelegram 相同逻辑
func BindCurrentTelegram(c *gin.Context) {
	BindTelegram(c)
}

// ─────────────────── 手机验证码（存根） ───────────────────

// POST /api/auth/phone-verification/send
func SendPhoneVerification(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
}

// POST /api/auth/phone-verification/verify
func VerifyPhoneCode(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "验证成功"})
}

// POST /api/auth/phone-verification/telegram-verify
func TelegramVerifyPhone(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "验证成功"})
}

// ─────────────────── 删除用户（管理员） ───────────────────

// DELETE /api/auth/users/:id
func DeleteUser(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	currentUser := middleware.CurrentUser(c)
	if currentUser.ID == id {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "不能删除自己"})
		return
	}
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "用户不存在"})
		return
	}
	if user.IsSuperAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "不能删除超级管理员"})
		return
	}
	database.DB.Model(&user).Update("is_active", false)
	c.JSON(http.StatusOK, gin.H{"message": "用户已停用"})
}

// POST /api/auth/users/:id/super-admin — 同 PATCH，兼容前端 POST 调用
func SetSuperAdminPost(c *gin.Context) {
	UpdateSuperAdmin(c)
}
