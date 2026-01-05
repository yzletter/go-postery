package ports

import "errors"

type EmailManager interface {
	Send(to string, code string) error
}

type VerifyEmailData struct {
	AppName   string
	Email     string
	Code      string
	ExpireMin int
	Year      int
	Address   string
}

var (
	ErrRenderEmailHTMLFailed = errors.New("渲染邮件 HTML 失败")
	ErrSendEmailFailed       = errors.New("发送邮件失败")
)
