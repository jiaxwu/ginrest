package ginrest

import (
	"context"

	"github.com/gin-gonic/gin"
)

// 服务函数
type ServiceOptFunc[Req any, Rsp any, Opt any] func(context.Context, *Req, ...Opt) (*Rsp, error)

// 服务函数
type ServiceFunc[Req any, Rsp any] func(context.Context, *Req) (*Rsp, error)

// 请求函数，对请求进一步处理的函数
type ReqFunc[Req any] func(c *gin.Context, req *Req)

// 处理请求
func DoOpt[Req any, Rsp any, Opt any](reqFunc ReqFunc[Req], serviceOptFunc ServiceOptFunc[Req, Rsp, Opt], opts ...Opt) gin.HandlerFunc {
	return do(reqFunc, nil, serviceOptFunc, opts...)
}

// 处理请求
func Do[Req any, Rsp any](reqFunc ReqFunc[Req], serviceFunc ServiceFunc[Req, Rsp]) gin.HandlerFunc {
	return do[Req, Rsp, struct{}](reqFunc, serviceFunc, nil)
}

// 处理请求
func do[Req any, Rsp any, Opt any](reqFunc ReqFunc[Req],
	serviceFunc ServiceFunc[Req, Rsp], serviceOptFunc ServiceOptFunc[Req, Rsp, Opt], opts ...Opt) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 参数绑定
		req, err := BindJSON[Req](c)
		if err != nil {
			return
		}
		// 进一步处理请求结构体
		if reqFunc != nil {
			reqFunc(c, req)
		}
		var rsp *Rsp
		// 业务逻辑函数调用
		if serviceFunc != nil {
			rsp, err = serviceFunc(c, req)
		} else if serviceOptFunc != nil {
			rsp, err = serviceOptFunc(c, req, opts...)
		} else {
			panic("must set ServiceFunc or ServiceFuncOpt")
		}
		// 处理响应
		ProcessRsp(c, rsp, err)
	}
}

// 参数绑定
func BindJSON[T any](c *gin.Context) (*T, error) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		FailureCodeMsg(c, ErrCodeInvalidReq, "invalid param")
		return nil, err
	}
	return &req, nil
}

// 设置上下文
func SetKey[T any](c *gin.Context, key string, value T) {
	c.Set(key, value)
}

// 获取上下文
func GetKey[T any](c *gin.Context, key string) T {
	var t T
	value, exists := c.Get(key)
	if !exists {
		return t
	}
	v, ok := value.(T)
	if !ok {
		panic("invald type")
	}
	return v
}
