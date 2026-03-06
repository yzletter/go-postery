package conf

// 短信验证码
const (
	SendSMSInterval = 60            // 发送间隔
	SMSValidTime    = 300           // 有效时间
	PhoneCodePrefix = "phone:code:" // 前缀
)

// 邮箱验证码
const (
	SendEmailInterval = 60
	EmailValidTime    = 600
	EmailCodePrefix   = "email:code:"
)
