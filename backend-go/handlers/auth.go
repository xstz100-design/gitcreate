package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"wholesale/config"
	"wholesale/database"
	"wholesale/middleware"
	"wholesale/models"
	"wholesale/services"
	"wholesale/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ─────────────────── OTP 内存存储 ───────────────────

type otpEntry struct {
	code      string
	userID    int64
	expiresAt time.Time
}

var (
	otpMu    sync.Mutex
	otpStore = make(map[int64]*otpEntry) // key: userID
)

// ─────────────────── Bot 深链登录 Token 存储 ───────────────────

type botLoginEntry struct {
	userID    *int64
	expiresAt time.Time
	confirmed bool
}

var botLoginStore sync.Map // key: token(string) → *botLoginEntry

// ConfirmBotLoginToken 供 Bot 消息处理器调用，将 token 与用户绑定
func ConfirmBotLoginToken(token string, userID int64) bool {
	val, ok := botLoginStore.Load(token)
	if !ok {
		return false
	}
	entry := val.(*botLoginEntry)
	if time.Now().After(entry.expiresAt) {
		botLoginStore.Delete(token)
		return false
	}
	entry.confirmed = true
	entry.userID = &userID
	return true
}

// POST /api/auth/bot-login/create — 生成深链 token，返回 bot_url
func BotLoginCreate(c *gin.Context) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "生成 Token 失败"})
		return
	}
	token := hex.EncodeToString(b)
	entry := &botLoginEntry{expiresAt: time.Now().Add(10 * time.Minute)}
	botLoginStore.Store(token, entry)

	botURL := "https://t.me/" + config.C.TGBotUsername + "?start=login_" + token
	c.JSON(http.StatusOK, gin.H{"token": token, "bot_url": botURL})
}

// GET /api/auth/bot-login/verify?token=TOKEN — 轮询确认结果
func BotLoginVerify(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "缺少 token"})
		return
	}
	val, ok := botLoginStore.Load(token)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"detail": "token 不存在"})
		return
	}
	entry := val.(*botLoginEntry)
	if time.Now().After(entry.expiresAt) {
		botLoginStore.Delete(token)
		c.JSON(http.StatusGone, gin.H{"detail": "token 已过期"})
		return
	}
	if !entry.confirmed || entry.userID == nil {
		// 202: 尚未确认，前端继续轮询
		c.JSON(http.StatusAccepted, gin.H{"detail": "等待用户确认"})
		return
	}
	var user models.User
	if err := database.DB.First(&user, *entry.userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "用户不存在"})
		return
	}
	botLoginStore.Delete(token) // 防止重复使用
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"detail": "账号已被禁用"})
		return
	}
	jwtToken, err := utils.CreateAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "生成 Token 失败"})
		return
	}
	c.JSON(http.StatusOK, TokenResponse{AccessToken: jwtToken, TokenType: "bearer", User: buildUserResponse(&user)})
}

func generateOTPCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}

var phoneRegexp = regexp.MustCompile(`^[+\d][\d\s\-]{5,19}$`)

// cambodiaPhoneRegexp 柬埔寨手机号：0XX-XXXXXX~XXXXXXX 或 +855XX-XXXXXX~XXXXXXX
// 支持 9~10 位本地格式 (0XXXXXXXXX) 及含国码格式 (+855XXXXXXXXX / 855XXXXXXXXX)
var cambodiaPhoneRegexp = regexp.MustCompile(`^(\+?855|0)[1-9]\d{7,8}$`)

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
// 仅允许管理员账号使用密码登录，普通用户请使用 OTP 或 Telegram
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

	// 非管理员账号不允许密码登录，引导使用 OTP
	if user.Role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"detail": "请使用 Telegram 验证码登录"})
		return
	}

	if !utils.VerifyPassword(password, user.HashedPassword) {
		middleware.LoginGuard.RecordFailure(ip, username)
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "用户名或密码错误"})
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
		// 全新用户：自动创建商户账号
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
			ApprovalStatus:     models.ApprovalApproved,
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
		phone := strings.TrimSpace(*req.Phone)
		if phone != "" && !cambodiaPhoneRegexp.MatchString(phone) {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "手机号格式不正确，请输入柬埔寨格式（如 012345678 或 +85512345678）"})
			return
		}
		updates["phone"] = phone
	}
	if req.Address != nil {
		addr := strings.TrimSpace(*req.Address)
		if addr != "" && len(addr) < 5 {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "地址太短，请填写准确地址"})
			return
		}
		updates["address"] = addr
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
	q := database.DB.Model(&models.User{})
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

// buildSortedJSON 将 map 按 key 字母序序列化为 JSON（用于 requestContact hash 校验）
func buildSortedJSON(m map[string]interface{}) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%q", k))
		sb.WriteString(":")
		switch v := m[k].(type) {
		case string:
			sb.WriteString(fmt.Sprintf("%q", v))
		case float64:
			sb.WriteString(strconv.FormatInt(int64(v), 10))
		case int64:
			sb.WriteString(strconv.FormatInt(v, 10))
		default:
			b, _ := json.Marshal(v)
			sb.Write(b)
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// verifyContactData 验证 requestContact 返回数据的 HMAC 签名
// 返回 (phone, tgUserID, error)
func verifyContactData(contactData map[string]interface{}) (string, int64, error) {
	token := config.C.TGBotToken
	if token == "" {
		return "", 0, fmt.Errorf("TG_BOT_TOKEN 未配置")
	}

	receivedHash, _ := contactData["hash"].(string)
	if receivedHash == "" {
		return "", 0, fmt.Errorf("缺少 hash 字段")
	}

	contactObj, ok := contactData["contact"].(map[string]interface{})
	if !ok {
		return "", 0, fmt.Errorf("缺少 contact 字段")
	}

	// 构造 data-check-string：除 hash 外的字段按字母序排列
	// contact 字段的值是按字母序序列化的 JSON
	var pairs []string
	for k, v := range contactData {
		if k == "hash" {
			continue
		}
		var valueStr string
		if k == "contact" {
			valueStr = buildSortedJSON(contactObj)
		} else {
			switch val := v.(type) {
			case string:
				valueStr = val
			case float64:
				valueStr = strconv.FormatInt(int64(val), 10)
			}
		}
		pairs = append(pairs, k+"="+valueStr)
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretKey := hmacSHA256([]byte("WebAppData"), []byte(token))
	computedHash := hex.EncodeToString(hmacSHA256(secretKey, []byte(dataCheckString)))

	if !hmac.Equal([]byte(computedHash), []byte(receivedHash)) {
		return "", 0, fmt.Errorf("签名验证失败")
	}

	// 验证有效期（5 分钟）
	if authDate, ok := contactData["auth_date"].(float64); ok {
		if time.Now().Unix()-int64(authDate) > 300 {
			return "", 0, fmt.Errorf("联系人授权已过期")
		}
	}

	phone, _ := contactObj["phone_number"].(string)
	var userID int64
	if uid, ok := contactObj["user_id"].(float64); ok {
		userID = int64(uid)
	}

	return phone, userID, nil
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
	// 默认最近7天
	days := 7
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 90 {
			days = v
		}
	}
	metrics := services.GetMetrics(days)
	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
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

// POST /api/auth/phone-verification/verify (需登录)
func VerifyPhoneCode(c *gin.Context) {
	user := middleware.CurrentUser(c)
	var req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "请输入手机号"})
		return
	}
	database.DB.Model(user).Update("phone", phone)
	database.DB.First(user, user.ID)
	c.JSON(http.StatusOK, buildUserResponse(user))
}

// POST /api/auth/phone-verification/telegram-verify (需登录)
func TelegramVerifyPhone(c *gin.Context) {
	user := middleware.CurrentUser(c)
	var req struct {
		Phone    string `json:"phone" binding:"required"`
		InitData string `json:"init_data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "请输入手机号"})
		return
	}
	// 若提供了 init_data，校验签名（可选加强验证）
	if req.InitData != "" {
		if _, err := validateTelegramInitData(req.InitData); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Telegram 验证失败: " + err.Error()})
			return
		}
	}
	database.DB.Model(user).Update("phone", phone)
	database.DB.First(user, user.ID)
	c.JSON(http.StatusOK, buildUserResponse(user))
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

	// 检查是否有未删除的订单
	var orderCount int64
	database.DB.Model(&models.Order{}).Where("merchant_id = ? AND is_deleted = ?", id, false).Count(&orderCount)
	if orderCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": fmt.Sprintf("该用户还有 %d 笔订单记录，请先在订单管理中处理后再删除", orderCount)})
		return
	}

	// 硬删除：清理关联数据后真正删除用户
	tx := database.DB.Begin()
	// 删除已删除的订单（软删除的历史订单）
	tx.Where("merchant_id = ?", id).Delete(&models.Order{})
	// 删除手机验证记录
	tx.Where("user_id = ?", id).Delete(&models.PhoneVerification{})
	// 删除用户本体
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "删除用户失败"})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}

// POST /api/auth/users/:id/super-admin — 同 PATCH，兼容前端 POST 调用
func SetSuperAdminPost(c *gin.Context) {
	UpdateSuperAdmin(c)
}

// ─────────────────── Telegram 验证码登录（浏览器端） ───────────────────

// POST /api/auth/otp/request
// 用手机号请求 Telegram 验证码，验证码通过 Bot 私聊发送
func RequestOTP(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	phone := strings.TrimSpace(req.Phone)

	botURL := "https://t.me/" + config.C.TGBotUsername

	var user models.User
	if err := database.DB.Where("username = ? OR phone = ?", phone, phone).First(&user).Error; err != nil {
		// 手机号未注册：引导用户先通过 Bot 注册账号
		c.JSON(http.StatusBadRequest, gin.H{
			"detail":   "该手机号未注册，请先通过 Telegram Bot 注册账号",
			"need_bot": true,
			"bot_url":  botURL,
		})
		return
	}

	if user.TelegramID == nil {
		// 账号存在但未绑定 Telegram：引导用户关联 Bot
		c.JSON(http.StatusBadRequest, gin.H{
			"detail":   "该账号未绑定 Telegram，请先通过 Bot 发送 /start 绑定账号",
			"need_bot": true,
			"bot_url":  botURL,
		})
		return
	}
	if !user.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "账号已被禁用"})
		return
	}

	code := generateOTPCode()
	otpMu.Lock()
	otpStore[user.ID] = &otpEntry{code: code, userID: user.ID, expiresAt: time.Now().Add(5 * time.Minute)}
	otpMu.Unlock()

	token := config.C.TGBotToken
	chatID := fmt.Sprintf("%d", *user.TelegramID)
	msg := fmt.Sprintf("🔐 <b>登录验证码：%s</b>\n\n有效期 5 分钟，请勿泄露给他人。", code)
	go utils.SendTelegramMessage(token, chatID, msg, nil)

	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送至您的 Telegram"})
}

// POST /api/auth/otp/verify
// 验证 Telegram 验证码并登录
func VerifyOTP(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ? OR phone = ?", strings.TrimSpace(req.Phone), strings.TrimSpace(req.Phone)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "验证码错误或已过期"})
		return
	}

	otpMu.Lock()
	entry, ok := otpStore[user.ID]
	if ok && time.Now().After(entry.expiresAt) {
		delete(otpStore, user.ID)
		ok = false
	}
	valid := ok && entry.code == req.Code
	if valid {
		delete(otpStore, user.ID)
	}
	otpMu.Unlock()

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "验证码错误或已过期"})
		return
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
	c.JSON(http.StatusOK, TokenResponse{AccessToken: token, TokenType: "bearer", User: buildUserResponse(&user)})
}

// ─────────────────── Telegram Login Widget（浏览器第三方登录） ───────────────────

// POST /api/auth/telegram-widget-login
// 接收浏览器端 Telegram Login Widget 回调数据，验证签名后登录
func TelegramWidgetLogin(c *gin.Context) {
	var req struct {
		ID        int64  `json:"id" binding:"required"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
		PhotoURL  string `json:"photo_url"`
		AuthDate  int64  `json:"auth_date" binding:"required"`
		Hash      string `json:"hash" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	// 构建 data-check-string（按字母顺序排序的 key=value，不含 hash）
	var pairs []string
	pairs = append(pairs, fmt.Sprintf("auth_date=%d", req.AuthDate))
	if req.FirstName != "" {
		pairs = append(pairs, fmt.Sprintf("first_name=%s", req.FirstName))
	}
	pairs = append(pairs, fmt.Sprintf("id=%d", req.ID))
	if req.LastName != "" {
		pairs = append(pairs, fmt.Sprintf("last_name=%s", req.LastName))
	}
	if req.PhotoURL != "" {
		pairs = append(pairs, fmt.Sprintf("photo_url=%s", req.PhotoURL))
	}
	if req.Username != "" {
		pairs = append(pairs, fmt.Sprintf("username=%s", req.Username))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// 验证签名：secretKey = SHA256(bot_token)，hash = HMAC-SHA256(data-check-string, secretKey)
	h := sha256.New()
	h.Write([]byte(config.C.TGBotToken))
	secretKey := h.Sum(nil)

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedHash), []byte(req.Hash)) {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Telegram 验证失败，签名不匹配"})
		return
	}

	// 检查 auth_date 是否在 24 小时内
	if time.Now().Unix()-req.AuthDate > 86400 {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "认证数据已过期，请重新登录"})
		return
	}

	// 按 telegram_id 查找用户
	var user models.User
	if err := database.DB.Where("telegram_id = ?", req.ID).First(&user).Error; err != nil {
		botURL := "https://t.me/" + config.C.TGBotUsername
		c.JSON(http.StatusBadRequest, gin.H{
			"detail":   "该 Telegram 账号未注册，请先通过 Bot 注册账号",
			"need_bot": true,
			"bot_url":  botURL,
		})
		return
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
	c.JSON(http.StatusOK, TokenResponse{AccessToken: token, TokenType: "bearer", User: buildUserResponse(&user)})
}

// ─────────────────── Telegram Mini App 关联已有账号 ───────────────────

// POST /api/auth/telegram-link-login
// 在 Mini App 中输入手机号+密码，将当前 Telegram 账号关联到已有的手机号账号
// 关联成功后，以后打开 Mini App 直接免密登录
func TelegramLinkLogin(c *gin.Context) {
	var req struct {
		InitData string `json:"init_data" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
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

	// 验证账号密码
	var user models.User
	if err := database.DB.Where("username = ?", strings.TrimSpace(req.Username)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "用户名或密码错误"})
		return
	}
	if !utils.VerifyPassword(req.Password, user.HashedPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "用户名或密码错误"})
		return
	}
	if !user.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "账号已被禁用"})
		return
	}

	// 若该 TG ID 已绑定另一个自动创建的 tg_xxx 账号，停用那个账号
	var existingTG models.User
	if database.DB.Where("telegram_id = ? AND id != ?", tgID, user.ID).First(&existingTG).Error == nil {
		database.DB.Model(&existingTG).Updates(map[string]interface{}{
			"is_active":   false,
			"telegram_id": nil,
		})
	}

	// 将 TG ID 绑定到当前账号
	database.DB.Model(&user).Update("telegram_id", tgID)
	database.DB.First(&user, user.ID)

	token, err := utils.CreateAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "生成 Token 失败"})
		return
	}
	c.JSON(http.StatusOK, TokenResponse{AccessToken: token, TokenType: "bearer", User: buildUserResponse(&user)})
}

// ─────────────────── Telegram requestContact 一键手机号关联 ───────────────────

// POST /api/auth/telegram-contact-link
// Mini App 中调用 Telegram.WebApp.requestContact() 后，将联系人数据发到后端
// 后端验证 HMAC 签名 + user_id 匹配，自动找到手机号账号并绑定
func TelegramContactLink(c *gin.Context) {
	var req struct {
		InitData    string                 `json:"init_data" binding:"required"`
		ContactData map[string]interface{} `json:"contact_data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	// 1. 验证 initData → 获取可信 tgID
	tgUser, err := validateTelegramInitData(req.InitData)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Telegram 验证失败: " + err.Error()})
		return
	}
	tgID := tgUser["id"].(int64)

	// 2. 验证 contact 数据 HMAC + 提取手机号和 user_id
	phone, contactUserID, err := verifyContactData(req.ContactData)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "联系人验证失败: " + err.Error()})
		return
	}

	// 3. 安全检查：contact.user_id 必须与 initData 的 tgID 完全一致
	//    防止用户伪造 phone_number 绑定他人账号
	if contactUserID != tgID {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "联系人与当前 Telegram 用户不匹配"})
		return
	}

	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "未获取到手机号"})
		return
	}

	// 4. 规范化手机号：去掉前缀 + 号，尝试两种格式查找账号
	phonePlain := strings.TrimPrefix(phone, "+")
	phoneWithPlus := "+" + phonePlain

	var user models.User
	if err := database.DB.Where("username = ? OR username = ?", phonePlain, phoneWithPlus).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "未找到该手机号对应的账号，请联系管理员"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "账号已被禁用"})
		return
	}

	// 5. 若该 TG ID 已绑定自动创建的 tg_xxx 账号，停用并解绑那个账号
	var existingTG models.User
	if database.DB.Where("telegram_id = ? AND id != ?", tgID, user.ID).First(&existingTG).Error == nil {
		database.DB.Model(&existingTG).Updates(map[string]interface{}{
			"is_active":   false,
			"telegram_id": nil,
		})
	}

	// 6. 绑定 TG ID 到手机号账号
	database.DB.Model(&user).Update("telegram_id", tgID)
	database.DB.First(&user, user.ID)

	jwtToken, err := utils.CreateAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "生成 Token 失败"})
		return
	}
	c.JSON(http.StatusOK, TokenResponse{AccessToken: jwtToken, TokenType: "bearer", User: buildUserResponse(&user)})
}
