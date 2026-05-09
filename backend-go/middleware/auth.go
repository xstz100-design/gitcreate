package middleware

import (
	"net/http"
	"strings"
	"wholesale/database"
	"wholesale/models"
	"wholesale/services"
	"wholesale/utils"

	"github.com/gin-gonic/gin"
)

const UserKey = "current_user"

// Auth JWT 鉴权中间件 - 从 Authorization: Bearer <token> 提取用户
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "未提供认证凭据"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "认证格式错误"})
			return
		}

		uid, err := utils.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "无法验证凭据"})
			return
		}

		var user models.User
		if err := database.DB.First(&user, uid).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "用户不存在"})
			return
		}

		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"detail": "账号已被禁用"})
			return
		}

		c.Set(UserKey, &user)
		// 活跃用户埋点：只统计商户/配货员，管理员不计（异步，不阻塞请求）
		if user.Role != models.RoleAdmin {
			go services.TrackActiveUser(user.ID)
		}
		c.Next()
	}
}

// RequireAdmin 要求管理员角色
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := CurrentUser(c)
		if user == nil || user.Role != models.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "需要管理员权限"})
			return
		}
		c.Next()
	}
}

// RequirePickerOrAdmin 配货员或管理员均可访问
func RequirePickerOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := CurrentUser(c)
		if user == nil || (user.Role != models.RoleAdmin && user.Role != models.RolePicker) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "需要配货员或管理员权限"})
			return
		}
		c.Next()
	}
}

// RequireSuperAdmin 要求超级管理员
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := CurrentUser(c)
		if user == nil || !user.IsSuperAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "需要超级管理员权限"})
			return
		}
		c.Next()
	}
}

// CurrentUser 从 gin.Context 取出当前用户
func CurrentUser(c *gin.Context) *models.User {
	v, exists := c.Get(UserKey)
	if !exists {
		return nil
	}
	user, _ := v.(*models.User)
	return user
}
