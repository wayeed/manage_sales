package middleware

import (
	"furniture-commission/internal/handler"

	"github.com/gin-gonic/gin"
)

// RBAC RBAC权限中间件（兼容旧代码，基于整数角色）
// roles: 允许访问的角色列表，空列表表示所有已认证用户都可访问
func RBAC(roles ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果没有指定角色限制，则所有已认证用户都可访问
		if len(roles) == 0 {
			c.Next()
			return
		}

		// 新逻辑：基于 role_codes 判断
		userRoleCodes := GetRoleCodes(c)
		if len(userRoleCodes) > 0 {
			// 有任意角色即可通过（兼容旧的 RBAC(5) 调用）
			c.Next()
			return
		}

		handler.Error(c, 403, "权限不足1")
		c.Abort()
	}
}

// RequireRole 基于角色编码的权限中间件
// roleCodes: 允许访问的角色编码列表，空列表表示所有已认证用户都可访问
func RequireRole(roleCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(roleCodes) == 0 {
			c.Next()
			return
		}

		userRoleCodes := GetRoleCodes(c)
		if len(userRoleCodes) == 0 {
			handler.Error(c, 403, "权限不足，请联系管理员分配角色")
			c.Abort()
			return
		}

		// 构建快速查找 map
		allowed := make(map[string]bool, len(roleCodes))
		for _, code := range roleCodes {
			allowed[code] = true
		}

		for _, code := range userRoleCodes {
			if allowed[code] {
				c.Next()
				return
			}
		}

		handler.Error(c, 403, "权限不足2")
		c.Abort()
	}
}

// RequireAdmin 仅管理员可访问（基于角色编码 "admin"）
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// RequirePermission 基于权限编码的权限中间件
// permissionCodes: 需要的权限编码列表，用户拥有其中任意一个即可通过
func RequirePermission(permissionCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(permissionCodes) == 0 {
			c.Next()
			return
		}

		userPermCodes := GetPermissions(c)
		if len(userPermCodes) == 0 {
			handler.Error(c, 403, "权限不足，请联系管理员分配权限")
			c.Abort()
			return
		}

		// 构建快速查找 set
		userPermSet := make(map[string]bool, len(userPermCodes))
		for _, code := range userPermCodes {
			userPermSet[code] = true
		}

		// 检查是否有所需权限之一
		for _, required := range permissionCodes {
			if userPermSet[required] {
				c.Next()
				return
			}
		}

		handler.Error(c, 403, "权限不足3")
		c.Abort()
	}
}

// RequireAllPermissions 要求拥有所有指定权限
func RequireAllPermissions(permissionCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(permissionCodes) == 0 {
			c.Next()
			return
		}

		userPermCodes := GetPermissions(c)
		userPermSet := make(map[string]bool, len(userPermCodes))
		for _, code := range userPermCodes {
			userPermSet[code] = true
		}

		for _, required := range permissionCodes {
			if !userPermSet[required] {
				handler.Error(c, 403, "权限不足4")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireRoleOrPermission 满足角色或权限其一即可通过
func RequireRoleOrPermission(roleCodes []string, permissionCodes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查角色
		if len(roleCodes) > 0 {
			userRoleCodes := GetRoleCodes(c)
			allowed := make(map[string]bool, len(roleCodes))
			for _, code := range roleCodes {
				allowed[code] = true
			}
			for _, code := range userRoleCodes {
				if allowed[code] {
					c.Next()
					return
				}
			}
		}

		// 检查权限
		if len(permissionCodes) > 0 {
			userPermCodes := GetPermissions(c)
			allowed := make(map[string]bool, len(permissionCodes))
			for _, code := range permissionCodes {
				allowed[code] = true
			}
			for _, code := range userPermCodes {
				if allowed[code] {
					c.Next()
					return
				}
			}
		}

		handler.Error(c, 403, "权限不足5")
		c.Abort()
	}
}
