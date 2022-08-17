package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jiaxwu/ginrest"
)

func main() {
	e := gin.New()
	e.Use(ginrest.Recovery())
	Register(e)
	if err := e.Run("127.0.0.1:8000"); err != nil {
		log.Println(err)
	}
}

// 注册请求
func Register(e *gin.Engine) {
	// 简单请求，不需要认证
	e.GET("/user/info/get", ginrest.Do(nil, GetUserInfo))
	// 认证，绑定UID，处理
	e.POST("/user/info/update", Verify, ginrest.Do(func(c *gin.Context, req *UpdateUserInfoReq) {
		req.UID = GetUID(c)
	}, UpdateUserInfo))
}

const (
	KeyUserID = "KeyUserID"
)

// 简单包装方便使用
func GetUID(c *gin.Context) int {
	return ginrest.GetKey[int](c, KeyUserID)
}

// 简单包装方便使用
func SetUID(c *gin.Context, uid int) {
	ginrest.SetKey(c, KeyUserID, uid)
}

// 认证
func Verify(c *gin.Context) {
	// 认证处理
	// ...
	// 忽略认证的具体逻辑
	SetUID(c, 10)
}

const (
	ErrCodeUserNotExists = 40100 // 用户不存在
)

type GetUserInfoReq struct {
	UID int `json:"uid"`
}

type GetUserInfoRsp struct {
	UID      int    `json:"uid"`
	Username string `json:"username"`
	Age      int    `json:"age"`
}

func GetUserInfo(ctx context.Context, req *GetUserInfoReq) (*GetUserInfoRsp, error) {
	if req.UID != 10 {
		return nil, ginrest.NewError(ErrCodeUserNotExists, "user not exists")
	}
	return &GetUserInfoRsp{
		UID:      req.UID,
		Username: "user_10",
		Age:      10,
	}, nil
}

type UpdateUserInfoReq struct {
	UID      int    `json:"uid"`
	Username string `json:"username"`
	Age      int    `json:"age"`
}

type UpdateUserInfoRsp struct{}

func UpdateUserInfo(ctx context.Context, req *UpdateUserInfoReq) (*UpdateUserInfoRsp, error) {
	if req.UID != 10 {
		return nil, ginrest.NewError(ErrCodeUserNotExists, "user not exists")
	}
	return &UpdateUserInfoRsp{}, nil
}
