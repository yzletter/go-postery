package interview

import "encoding/json"

// UploadCallbackRequest OSS 上传回调请求
type UploadCallbackRequest struct {
	Bucket string      `json:"bucket"`
	Size   json.Number `json:"size"`
	Object string      `json:"object"`
}
