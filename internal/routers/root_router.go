package routers

import (
	"go-api-testing/internal/api"

	"github.com/gin-gonic/gin"
)

// 根路由配置
func rootRouters(r *gin.RouterGroup) {
	r.GET("/index", api.Index)
}
