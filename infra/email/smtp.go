package email

import (
	"bytes"
	"html/template"
	"log/slog"

	"github.com/yzletter/go-postery/service/ports"
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

func NewEmailManager(from string, authCode string, subject string, expireMin int, appName string, year int, address string) ports.EmailManager {
	return &QQEmailSMTPManager{
		from:      from,
		authCode:  authCode,
		subject:   subject,
		expireMin: expireMin,
		appName:   appName,
		year:      year,
		address:   address,
	}
}

func (m *QQEmailSMTPManager) Send(to string, code string) error {
	data := ports.VerifyEmailData{
		AppName:   m.appName,
		Email:     to,
		Code:      code,
		ExpireMin: m.expireMin,
		Year:      m.year,
		Address:   m.address,
	}
	htmlBody, err := renderVerifyEmailHTML(data)
	if err != nil {
		slog.Error("Render Email HTML Failed", "error", err)
		return ports.ErrRenderEmailHTMLFailed
	}

	message := gomail.NewMessage()
	message.SetHeader("From", m.from)
	message.SetHeader("To", to)
	message.SetHeader("Subject", m.subject)
	message.SetBody("text/html", htmlBody) // 关键：HTML

	d := gomail.NewDialer("smtp.qq.com", 465, m.from, m.authCode)
	// d.TLSConfig = &tls.Config{ServerName: "smtp.qq.com"} // 如有需要可加

	if err := d.DialAndSend(message); err != nil {
		slog.Error("Send Email Failed", "error", err)
		return ports.ErrSendEmailFailed
	}
	return nil
}

func renderVerifyEmailHTML(data ports.VerifyEmailData) (string, error) {
	tpl, err := template.ParseFiles("./infra/email/verify_email.html")
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
