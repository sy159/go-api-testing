package api

import (
	"github.com/gin-gonic/gin"
	"go-api-testing/utils/response"
)

// Index 首页
func Index(c *gin.Context) {
	response.Success(c, gin.H{})
}
