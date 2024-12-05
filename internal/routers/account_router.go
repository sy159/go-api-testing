package routers

import (
	"github.com/gin-gonic/gin"
	. "go-api-testing/internal/api/account"
	"go-api-testing/internal/routers/middleware"
)

// 用户相关路由
func userRouters(r *gin.RouterGroup) {
	userGroup := r.Group("/account")
	//userGroup.Use(middleware.JWTAuth()) // 账户相关的必须要登录
	{
		userGroup.POST("/login", Login)
		userGroup.POST("/refresh_token", RefreshToken)
		userGroup.GET("/user", middleware.JWTAuth(), GetUserList)
		userGroup.POST("/user", CreateUser)
		userGroup.DELETE("/user", middleware.JWTAuth(), DelUser)
		userGroup.PUT("/user", middleware.JWTAuth(), UpdateUser)
	}
}
