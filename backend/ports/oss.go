package ports

type OSSManager interface {
	Sign(dir string, callbackURL string) (string, error) // Sign 返回服务端签名给前端上传
	Resign(objectName string) (string, error)            // Resign 返回预签名URL给前端下载
}
