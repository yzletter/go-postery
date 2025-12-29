package sms

import (
	"context"
	"fmt"
	"log/slog"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dypnsapi20170525 "github.com/alibabacloud-go/dypnsapi-20170525/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"github.com/yzletter/go-postery/conf"
	"github.com/yzletter/go-postery/service/ports"
)

type AliyunSmsClient struct {
	internalClient *dypnsapi20170525.Client
}

func NewAliyunSmsClient(AccessKeyId, AccessKeySecret string) ports.SmsClient {
	config := openapi.Config{
		AccessKeyId:     tea.String(AccessKeyId),
		AccessKeySecret: tea.String(AccessKeySecret),
		Endpoint:        tea.String("dypnsapi.aliyuncs.com"), // 不同服务域名不同
		Credential:      credential.Credential(nil),
	}

	client, err := dypnsapi20170525.NewClient(&config)
	if err != nil {
		slog.Error("Aliyun SMS Client New Failed", "error", err)
		return nil
	}

	return &AliyunSmsClient{
		internalClient: client,
	}
}

func (client *AliyunSmsClient) SendSms(ctx context.Context, phoneNumber string, code string) error {
	runtime := &util.RuntimeOptions{}

	req := &dypnsapi20170525.SendSmsVerifyCodeRequest{
		PhoneNumber:      tea.String(phoneNumber),
		SignName:         tea.String("速通互联验证码"),
		TemplateCode:     tea.String("100001"),
		TemplateParam:    tea.String(fmt.Sprintf("{\"code\":\"%s\",\"min\":\"5\"}", code)),
		CodeLength:       tea.Int64(6),                           // 验证码长度支持 4～8 位长度，默认是 4 位。
		ValidTime:        tea.Int64(int64(conf.SMSValidTime)),    // 验证码有效时长，单位秒，默认为 300 秒。
		Interval:         tea.Int64(int64(conf.SendSMSInterval)), // 时间间隔，单位：秒。即多久间隔可以发送一次验证码，用于频控，默认 60 秒。
		DuplicatePolicy:  tea.Int64(1),                           // 核验规则，当有效时间内对同场景内的同号码重复发送验证码时，旧验证码如何处理。 1 表示覆盖 2 表示都有效
		ReturnVerifyCode: tea.Bool(true),                         // 是否返回验证码
	}

	resp, err := client.internalClient.SendSmsVerifyCodeWithContext(ctx, req, runtime)
	if err != nil {
		slog.Error(err.Error())
		return ports.ErrSendSMSFailed
	}

	slog.Info(resp.Body.String())
	return nil
}

func (client *AliyunSmsClient) CheckSms(ctx context.Context, phoneNumber string, code string) error {
	runtime := &util.RuntimeOptions{}
	req := &dypnsapi20170525.CheckSmsVerifyCodeRequest{
		PhoneNumber: tea.String(phoneNumber),
		VerifyCode:  tea.String(code),
	}
	resp, err := client.internalClient.CheckSmsVerifyCodeWithContext(ctx, req, runtime)
	if err != nil {
		slog.Error(err.Error())
		return ports.ErrCheckSMSFailed
	}

	slog.Info(resp.Body.String())
	return nil
}
