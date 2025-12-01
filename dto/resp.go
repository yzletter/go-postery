package dto

// Resp 统一定义后端返回数据格式
type Resp struct {
	Code int         `json:"code"`           // 业务码
	Msg  string      `json:"msg"`            // 信息
	Data interface{} `json:"data,omitempty"` // 数据
}
