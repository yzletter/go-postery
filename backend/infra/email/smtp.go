package email

import (
	"bytes"
	"context"
	_ "embed"
	"html/template"
	"log/slog"

	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/ports"
	"gopkg.in/gomail.v2"
)

type QQEmailSMTPManager struct {
	from      string
	authCode  string // 授权码
	subject   string // 主题
	appName   string // 应用名称
	expireMin int    // 有效时间
	year      int    // 年份
	address   string // 公司地址
}

type VerifyEmailData struct {
	AppName   string
	Email     string
	Code      string
	ExpireMin int
	Year      int
	Address   string
}

func NewSMTPEmailClient(config conf.EmailConfig) ports.CodeClient {
	return &QQEmailSMTPManager{
		from:      config.From,
		authCode:  config.AuthCode,
		subject:   config.Subject,
		expireMin: config.ExpireMin,
		appName:   config.AppName,
		year:      config.Year,
		address:   config.Address,
	}
}

func (m *QQEmailSMTPManager) Send(ctx context.Context, identifier string, code string) error {
	data := VerifyEmailData{
		Email:     identifier,
		Code:      code,
		AppName:   m.appName,
		ExpireMin: m.expireMin,
		Year:      m.year,
		Address:   m.address,
	}
	htmlBody, err := renderVerifyEmailHTML(data)
	if err != nil {
		slog.Error("Render Email HTML Failed", "error", err)
		return ports.ErrSendCodeFailed
	}

	message := gomail.NewMessage()
	message.SetHeader("From", m.from)
	message.SetHeader("To", identifier)
	message.SetHeader("Subject", m.subject)
	message.SetBody("text/html", htmlBody) // 关键：HTML

	d := gomail.NewDialer("smtp.qq.com", 465, m.from, m.authCode)
	// d.TLSConfig = &tls.Config{ServerName: "smtp.qq.com"} // 如有需要可加

	if err := d.DialAndSend(message); err != nil {
		slog.Error("Send Email Failed", "error", err)
		return ports.ErrSendCodeFailed
	}
	return nil
}

//go:embed verify_email.html
var verify_template string

func renderVerifyEmailHTML(data VerifyEmailData) (string, error) {
	tpl, err := template.New("verify_email.html").Parse(verify_template)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
