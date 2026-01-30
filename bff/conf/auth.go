package conf

const (
	UserIDInContext = "user_id" // uid 在上下文中的 Name
	PhoneCodePrefix = "phone:code:"
	EmailCodePrefix = "email:code:"
)

type CodeBiz int

const (
	SMSCode CodeBiz = iota + 1
	EmailCode
)
