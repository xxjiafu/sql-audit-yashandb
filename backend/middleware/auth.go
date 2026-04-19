package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"sql-audit/services"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			c.Abort()
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := services.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireAdminRole 需要管理员、组长或DBA角色才能访问
func RequireAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			c.Abort()
			return
		}

		r := role.(string)
		// admin, leader, dba 都可以访问管理界面查看和审核工单
		if r != "admin" && r != "leader" && r != "dba" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问，需要管理员/组长/DBA角色"})
			c.Abort()
			return
		}

		c.Next()
	}
}
