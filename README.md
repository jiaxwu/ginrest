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
    - 对应Do()、DoOpt()方法
  - 封装标准格式服务处理结果为标准RESTful响应格式：Rsp{code, msg, data}
    - 对应ProcessRsp()、Success()、Failure()、FailureCodeMsg()方法
  - 默认使用统一数字错误码格式：\[0, 4XXXX, 5XXXX\]
  - 默认使用标准错误格式：Error{code, msg}
    - 对应NewError()、ToError()、ErrCode()、ErrMsg()、ErrEqual()方法
  - 默认统一状态码\[200, 400, 500\]
    - 对应错误码\[0, 4XXXX, 5XXXX\]
- 提供Recovery中间件，统一panic时的响应格式
- 提供SetKey()、GetKey()方法，用于存储请求上下文（泛型）
- 提供ReqFunc()，用于设置Req（泛型）

# 使用例子

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

# 注
如果这个库的设计不符合具体的业务，也可以按照这种思路去封装一个类似的库，只要尽可能的统一请求、响应的格式，就可以减少很多重复的模板代码。