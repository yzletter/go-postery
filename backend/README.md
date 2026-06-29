# 后端

## 文件目录结构

```
├── README.md
├── conf		// 微服务配置
├── errs		// 微服务错误
├── event		// 消息
├── grpc		// grpc 相关
├── infra 	// 基层设施
├── micro		// 微服务模块
├── pkg			// 通用包
├── ports		// 微服务调用接口
├── script	// 脚本
└── utils		// 小工具
```



## 微服务

### DAO

### 返回的错误

```go
var (
	ErrServerInternal = errors.New("数据库内部错误")
	ErrRecordNotFound = errors.New("记录不存在")
	ErrUniqueKey      = errors.New("唯一键冲突")
	ErrParamsInvalid  = errors.New("参数有误")
)
```

### Cache