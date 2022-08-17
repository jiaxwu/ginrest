package ginrest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 响应
type Rsp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// 处理简单响应
func ProcessRsp(c *gin.Context, rsp any, err error) {
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, rsp)
}

// 请求成功
func Success(c *gin.Context, data any) {
	ginRsp(c, http.StatusOK, &Rsp{
		Code: ErrCodeOK,
		Msg:  "ok",
		Data: data,
	})
}

// 请求失败
func Failure(c *gin.Context, err error) {
	var rsp Rsp
	if e, ok := err.(*Error); ok {
		rsp.Code = e.Code
		rsp.Msg = e.Msg
	} else {
		rsp.Code = ErrCodeUnknownException
		rsp.Msg = "unknwon exception"
	}
	ginRsp(c, getHttpStatusByCode(rsp.Code), &rsp)
	c.Abort()
}

// 请求失败
func FailureCodeMsg(c *gin.Context, code int, msg string) {
	ginRsp(c, getHttpStatusByCode(code), &Rsp{
		Code: code,
		Msg:  msg,
	})
	c.Abort()
}

// gin响应
func ginRsp(c *gin.Context, code int, rsp any) {
	c.JSON(code, rsp)
}

// 根据错误码获取HttpStatus
func getHttpStatusByCode(code int) int {
	if (code / 10000) == 4 {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
