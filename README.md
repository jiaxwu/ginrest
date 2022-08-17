[TOC]

# ginrest
消除RESTful接口的模板代码

# 背景
基于现在微服务或者服务化的思想，我们大部分的业务逻辑处理函数都是长这样的：

比如grpc服务端：
```go
func (s *Service) GetUserInfo(ctx context.Context, req *pb.GetUserInfoReq) (*pb.GetUserInfoRsp, error) {
    // 业务逻辑
    // ...
}
```
grpc客户端：
```go
func (s *Service) GetUserInfo(ctx context.Context, req *pb.GetUserInfoReq, opts ...grpc.CallOption) (*pb.GetUserInfoRsp, error) {
    // 业务逻辑
    // ...
}
```

有些服务我们需要把它包装为RESTful形式的接口，一般需要经历以下步骤：
1. 指定HTTP方法、URL
2. **鉴权**
3. **参数绑定**
4. 处理请求
5. **处理响应**

可以发现，参数绑定、处理响应几乎都是一样模板代码，鉴权也基本上是模板代码（当然有些鉴权可能比较复杂）。

而**Ginrest库就是为了消除这些模板代码**，它不是一个复杂的框架，只是一个简单的库，辅助处理这些重复的事情，为了实现这个能力使用了Go1.18的泛型。

仓库地址：https://github.com/jiaxwu/ginrest

# 特性
这个库提供以下特性：
- 封装RESTful请求响应
  - 封装RESTful请求为标准格式服务
  - 封装标准格式服务处理结果为标准RESTful响应格式：Rsp{code, msg, data}
  - 默认使用统一数字错误码格式：\[0, 4XXXX, 5XXXX\]
  - 默认使用标准错误格式：Error{code, msg}
  - 默认统一状态码\[200, 400, 500\]
- 提供Recovery中间件，统一panic时的响应格式
- 提供SetKey()、GetKey()方法，用于存储请求上下文（泛型）
- 提供ReqFunc()，用于设置Req（泛型）

# 使用例子
示例代码在：https://github.com/jiaxwu/ginrest/blob/main/examples/main.go

首先我们实现两个简单的服务：
```go
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
```
然后使用Gin+Ginrest包装为RESTful接口：

可以看到Register()里面每个接口都只需要一行代码！
```go
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
        reqFunc := func(c *gin.Context, req *UpdateUserInfoReq) {
		req.UID = GetUID(c)
	} // 这里拆多一步是为了显示第一个参数是ReqFunc
	e.POST("/user/info/update", Verify, ginrest.Do(reqFunc, UpdateUserInfo))
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
```


运行上面代码，然后尝试访问接口，可以看到返回结果：
```
请求1
GET http://127.0.0.1:8000/user/info/get
{
    "uid": 10
}
响应1
{
    "code": 0,
    "msg": "ok",
    "data": {
        "uid": 10,
        "username": "user_10",
        "age": 10
    }
}

请求2
GET http://127.0.0.1:8000/user/info/get
{
    "uid": 1
}
响应2
{
    "code": 40100,
    "msg": "user not exists"
}

请求3
POST http://127.0.0.1:8000/user/info/update
{
    "username": "jiaxwu",
    "age": 10
}
响应3
{
    "code": 0,
    "msg": "ok",
    "data": {}
}
```

# 实现原理
Do()和DoOpt()都会转发到do()，它其实是一个模板函数，把脏活累活给处理了：
```go
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
```

# 功能列表

## 处理请求
用于把一个标准服务封装为一个RESTful`gin.HandlerFunc`，对应Do()、DoOpt()函数。

DoOpt()相比于Do()多了一个opts参数，因为很多rpc框架客户端都有一个opts参数作为结尾。

还有一个`BindJSON()`，用于把请求体包装为一个Req结构体：
```go
// 参数绑定
func BindJSON[T any](c *gin.Context) (*T, error) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		FailureCodeMsg(c, ErrCodeInvalidReq, "invalid param")
		return nil, err
	}
	return &req, nil
}
```
如果无法使用Do()和DoOpt()则可以使用此方法。

## 处理响应
用于把rsp、error、errcode、errmsg等数据封装为一个JSON格式响应体，对应ProcessRsp()、Success()、Failure()、FailureCodeMsg()函数。

比如`ProcessRsp()`需要带上rsp和error，这样业务里面就不需要再写如下模板代码了：
```go
// 处理简单响应
func ProcessRsp(c *gin.Context, rsp any, err error) {
	if err != nil {
		Failure(c, err)
		return
	}
	Success(c, rsp)
}
```

响应格式统一为：
```go
// 响应
type Rsp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}
```

`Success()`用于处理成功情况：
```go
// 请求成功
func Success(c *gin.Context, data any) {
	ginRsp(c, http.StatusOK, &Rsp{
		Code: ErrCodeOK,
		Msg:  "ok",
		Data: data,
	})
}
```
其余同理。

如果无法使用Do()和DoOpt()则可以使用这些方法。

## 处理错误
一般我们都需要在出错时带上一个业务错误码，方便客户端处理。因此我们需要提供一个合适的error类型：
```go
// 错误
type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
```
我们提供了一些函数方便使用`Error`，对应NewError()、ToError()、ErrCode()、ErrMsg()、ErrEqual()函数。

比如`NewError()`生成一个Error类型error：
```go
// 通过code和msg产生一个错误
func NewError(code int, msg string) error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}
```

## 请求上下文操作
Gin的请求是链式处理的，也就是多个handler顺序的处理一个请求，比如：
```go
        reqFunc := func(c *gin.Context, req *UpdateUserInfoReq) {
		req.UID = ginrest.GetKey[int](c, KeyUserID)
	}
        // 认证，绑定UID，处理
	e.POST("/user/info/update", Verify, ginrest.Do(reqFunc, UpdateUserInfo))
```
这个接口经历了Verify和ginrest.Do两个handler，其中我们在Verify的时候通过认证知道了用户的身份信息（比如uid），我们希望把这个uid存起来，这样可以在业务逻辑里使用。

因此我们提供了SetKey()、GetKey()两个函数，用于存储请求上下文：

比如认证通过后我们可以设置UID到上下文，然后在reqFunc()里读取设置到req里面（下面介绍）。
```go
// 认证
func Verify(c *gin.Context) {
	// 认证处理
	// ...
	// 忽略认证的具体逻辑
	ginrest.SetKey(c, KeyUserID, uid)
}
```

## 请求结构体处理
上面我们设置了请求上下文，比如UID，但是其实我们并不知道具体这个UID是需要设置到req里的哪个字段，因此我们提供了一个回调函数ReqFunc()，用于设置Req：
```go
	// 这里↓
        reqFunc := func(c *gin.Context, req *UpdateUserInfoReq) {
		req.UID = ginrest.GetKey[int](c, KeyUserID)
	}
        // 认证，绑定UID，处理
	e.POST("/user/info/update", Verify, ginrest.Do(reqFunc, UpdateUserInfo))
```

# 注
如果这个库的设计不符合具体的业务，也可以按照这种思路去封装一个类似的库，只要尽可能的统一请求、响应的格式，就可以减少很多重复的模板代码。