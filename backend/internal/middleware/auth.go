package middleware

import (
	"strings"

	"furniture-commission/internal/handler"
	appauthcache "furniture-commission/internal/pkg/authcache"
	appjwt "furniture-commission/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	ContextUserID      = "user_id"
	ContextRole        = "role"
	ContextRoleCodes   = "role_codes"
	ContextPermissions = "permissions"
	ContextStoreID     = "storeID"
)

// SetDB 设置数据库实例（在路由初始化时调用）
func SetDB(db *gorm.DB) {
	_authDB = db
}

var _authDB *gorm.DB

// Auth JWT认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			handler.Error(c, 401, "缺少认证令牌")
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			handler.Error(c, 401, "认证令牌格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := appjwt.ParseToken(tokenString)
		if err != nil {
			handler.Error(c, 401, "认证令牌无效或已过期")
			c.Abort()
			return
		}

		// 将用户ID存入上下文
		c.Set(ContextUserID, claims.UserID)

		// 从缓存或数据库查询用户角色编码
		roleCodes := getUserRoleCodes(claims.UserID)
		c.Set(ContextRoleCodes, roleCodes)

		// 从缓存或数据库查询用户权限编码（完整链路: user_roles → roles → role_permissions → permissions）
		permCodes := getUserPermissionCodes(claims.UserID)
		c.Set(ContextPermissions, permCodes)

		// 兼容：设置 role 为 1（表示已认证），保持旧代码不报错
		c.Set(ContextRole, 1)

		// 查询用户的门店ID并设置到上下文
		if _authDB != nil {
			var storeID int64
			_authDB.Table("users").Select("store_id").Where("id = ?", claims.UserID).Scan(&storeID)
			if storeID > 0 {
				c.Set(ContextStoreID, storeID)
			}
		}

		c.Next()
	}
}

// getUserRoleCodes 获取用户的角色编码
func getUserRoleCodes(userID int64) []string {
	// 查缓存
	if codes, ok := appauthcache.Cache.Get(userID); ok {
		return codes
	}

	// 查数据库
	if _authDB == nil {
		return []string{}
	}

	var roleCodes []string
	_authDB.Table("user_roles").
		Select("roles.role_code").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.status = 1", userID).
		Pluck("roles.role_code", &roleCodes)

	if roleCodes == nil {
		roleCodes = []string{}
	}

	// 写缓存
	appauthcache.Cache.Set(userID, roleCodes)

	return roleCodes
}

// getUserPermissionCodes 获取用户的权限编码
// 完整链路: user_roles → roles → role_permissions → permissions
func getUserPermissionCodes(userID int64) []string {
	// 查缓存
	if codes, ok := appauthcache.PermCache.GetPermissions(userID); ok {
		return codes
	}

	// 查数据库
	if _authDB == nil {
		return []string{}
	}

	var permCodes []string
	_authDB.Table("permissions").
		Select("DISTINCT permissions.permission_code").
		Joins("INNER JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Joins("INNER JOIN user_roles ur ON ur.role_id = rp.role_id").
		Joins("INNER JOIN roles r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND permissions.status = 1 AND r.status = 1", userID).
		Pluck("DISTINCT permissions.permission_code", &permCodes)

	if permCodes == nil {
		permCodes = []string{}
	}

	// 写缓存
	appauthcache.PermCache.SetPermissions(userID, permCodes)

	return permCodes
}

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) int64 {
	if userID, exists := c.Get(ContextUserID); exists {
		return userID.(int64)
	}
	return 0
}

// GetRole 从上下文中获取用户角色（兼容旧代码）
func GetRole(c *gin.Context) int {
	if role, exists := c.Get(ContextRole); exists {
		return role.(int)
	}
	return 0
}

// GetRoleCodes 从上下文中获取用户角色编码列表
func GetRoleCodes(c *gin.Context) []string {
	if codes, exists := c.Get(ContextRoleCodes); exists {
		if arr, ok := codes.([]string); ok {
			return arr
		}
	}
	return []string{}
}

// GetPermissions 从上下文中获取用户权限编码列表
func GetPermissions(c *gin.Context) []string {
	if perms, exists := c.Get(ContextPermissions); exists {
		if arr, ok := perms.([]string); ok {
			return arr
		}
	}
	return []string{}
}

// GetStoreID 从上下文中获取门店ID
func GetStoreID(c *gin.Context) int64 {
	if storeID, exists := c.Get(ContextStoreID); exists {
		if id, ok := storeID.(int64); ok {
			return id
		}
	}
	return 0
}
