package chatagent

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 模型列表数据
var modelList = []gin.H{
	// TODO: 添加Chat-Agent模型
}

// 获取模型列表
func handleListModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   modelList,
	})
}

// 获取单个模型
func handleGetModel(c *gin.Context) {
	modelID := c.Param("model")
	for _, m := range modelList {
		if m["id"] == modelID {
			c.JSON(http.StatusOK, m)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
}

// Chat completions 处理函数
func handleChatCompletions(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Chat-Agent module not implemented yet",
	})
}

// RegisterRoutes 注册Chat-Agent模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/models", handleListModels)
	r.GET("/models/:model", handleGetModel)
	r.POST("/chat/completions", handleChatCompletions)
}
