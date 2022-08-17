package ginrest

import (
	"log"

	"github.com/gin-gonic/gin"
)

// 异常恢复
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Println("a panic captured", err)
		FailureCodeMsg(c, ErrCodeUnknownException, "unknown exception")
	})
}
