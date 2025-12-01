package utils

import (
	"errors"
	"fmt"
	"strings"
)
import "github.com/go-playground/validator/v10"

// BindErrMsg 捕获 JSON 绑定结构体失败的具体信息
func BindErrMsg(err error) string {

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		msg := make([]string, 0)
		for _, validationErr := range validationErrs {
			msg = append(msg, fmt.Sprintf("字段 %s 不满足 %s", validationErr.Field(), validationErr.Tag()))
		}
		return strings.Join(msg, ";")
	}

	return ""
}
