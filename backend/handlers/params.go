package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	raw := c.Param(name)
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing " + name})
		return 0, false
	}
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || n == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	if n > uint64(^uint(0)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return uint(n), true
}
