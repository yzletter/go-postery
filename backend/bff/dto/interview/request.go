package interview

// UploadCallbackRequest OSS 上传回调请求
type UploadCallbackRequest struct {
	Bucket string `json:"bucket"`
	Size   string `json:"size"`
	Object string `json:"object"`
}
