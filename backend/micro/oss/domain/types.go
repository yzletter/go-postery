package domain

type UploadBiz int32

const (
	UploadBizUnspecified           UploadBiz = 0
	UploadBizUserAvatar            UploadBiz = 1
	UploadBizInterviewQuestionBank UploadBiz = 2
)

type CallbackResult struct {
	Biz        UploadBiz
	UserID     int64
	ObjectName string
	SourceFile string
	Size       string
}
