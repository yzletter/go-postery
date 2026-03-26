package oss

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"log/slog"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	sdk_credentials "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/aliyun/credentials-go/credentials"
	"github.com/yzletter/go-postery/microservice-backend/user/config"
	"github.com/yzletter/go-postery/microservice-backend/user/service/ports"
)

// 定义全局变量
var (
	region      string
	bucketName  string
	product     = "oss"
	callbackUrl string
)

// PolicyToken 结构体用于存储生成的表单数据
type PolicyToken struct {
	Policy           string `json:"policy"`
	SecurityToken    string `json:"security_token"`
	SignatureVersion string `json:"x_oss_signature_version"`
	Credential       string `json:"x_oss_credential"`
	Date             string `json:"x_oss_date"`
	Signature        string `json:"signature"`
	Host             string `json:"host"`
	Dir              string `json:"dir"`
	Callback         string `json:"callback"` // 回调
}

type CallbackParam struct {
	CallbackUrl      string `json:"callbackUrl"`
	CallbackBody     string `json:"callbackBody"`
	CallbackBodyType string `json:"callbackBodyType"`
}

type AliyunOSSManager struct {
	AccessKeyID     string
	AccessKeySecret string
	Arn             string
}

func Init(config config.OSSConfig) ports.OSSManager {
	region = "cn-hongkong"    // Bucket 所处地域
	bucketName = "go-postery" // Bucket 名称
	callbackUrl = "http://gopostery.top/api/v1/users/callback"

	return &AliyunOSSManager{
		AccessKeyID:     config.AccessKeyID,
		AccessKeySecret: config.AccessKeySecret,
		Arn:             config.Arn,
	}
}

func (manager *AliyunOSSManager) Sign(dir string) (string, error) {

	// 设置 OSS 上传地址
	host := fmt.Sprintf("https://%s.oss-%s.aliyuncs.com", bucketName, region)

	config := new(credentials.Config).
		SetType("ram_role_arn").
		SetAccessKeyId(manager.AccessKeyID). // 填写 AccessKeyID
		SetAccessKeySecret(manager.AccessKeySecret). // 填写 AccessKeySecret
		SetRoleArn(manager.Arn). // 填写 Arn
		SetRoleSessionName("Upload").
		SetPolicy("").
		SetRoleSessionExpiration(3600)

	// 根据配置创建凭证提供器
	provider, err := credentials.NewCredential(config)
	if err != nil {
		slog.Error("OSS NewCredential fail", "error", err)
		return "", err
	}

	// 从凭证提供器获取凭证
	cred, err := provider.GetCredential()
	if err != nil {
		slog.Error("OSS GetCredential fail", "error", err)
		return "", err
	}

	// 构建policy
	utcTime := time.Now().UTC()
	date := utcTime.Format("20060102")
	expiration := utcTime.Add(1 * time.Hour)
	policyMap := map[string]any{
		"expiration": expiration.Format("2006-01-02T15:04:05.000Z"),
		"conditions": []any{
			map[string]string{"bucket": bucketName},
			map[string]string{"x-oss-signature-version": "OSS4-HMAC-SHA256"},
			map[string]string{"x-oss-credential": fmt.Sprintf("%v/%v/%v/%v/aliyun_v4_request", *cred.AccessKeyId, date, region, product)},
			map[string]string{"x-oss-date": utcTime.Format("20060102T150405Z")},
			map[string]string{"x-oss-security-token": *cred.SecurityToken},
		},
	}

	// 将policy转换为 JSON 格式
	policy, err := json.Marshal(policyMap)
	if err != nil {
		slog.Error("OSS json.Marshal fail", "error", err)
	}

	// 构造待签名字符串（StringToSign）
	stringToSign := base64.StdEncoding.EncodeToString([]byte(policy))

	hmacHash := func() hash.Hash { return sha256.New() }
	// 构建signing key
	signingKey := "aliyun_v4" + *cred.AccessKeySecret
	h1 := hmac.New(hmacHash, []byte(signingKey))
	io.WriteString(h1, date)
	h1Key := h1.Sum(nil)

	h2 := hmac.New(hmacHash, h1Key)
	io.WriteString(h2, region)
	h2Key := h2.Sum(nil)

	h3 := hmac.New(hmacHash, h2Key)
	io.WriteString(h3, product)
	h3Key := h3.Sum(nil)

	h4 := hmac.New(hmacHash, h3Key)
	io.WriteString(h4, "aliyun_v4_request")
	h4Key := h4.Sum(nil)

	// 生成签名
	h := hmac.New(hmacHash, h4Key)
	io.WriteString(h, stringToSign)
	signature := hex.EncodeToString(h.Sum(nil))

	// 回调
	var callbackParam CallbackParam
	callbackParam.CallbackUrl = callbackUrl
	callbackParam.CallbackBody = "{\"bucket\":\"${bucket}\",\"object\":\"${object}\",\"size\":\"${size}\"}"
	callbackParam.CallbackBodyType = "application/json"
	callback_str, err := json.Marshal(callbackParam)
	if err != nil {
		fmt.Println("callback json err:", err)
	}
	callbackBase64 := base64.StdEncoding.EncodeToString(callback_str)

	// 构建返回给前端的表单
	policyToken := PolicyToken{
		Policy:           stringToSign,
		SecurityToken:    *cred.SecurityToken,
		SignatureVersion: "OSS4-HMAC-SHA256",
		Credential:       fmt.Sprintf("%v/%v/%v/%v/aliyun_v4_request", *cred.AccessKeyId, date, region, product),
		Date:             utcTime.UTC().Format("20060102T150405Z"),
		Signature:        signature,
		Host:             host,           // 返回 OSS 上传地址
		Dir:              dir,            // 返回上传目录
		Callback:         callbackBase64, // 返回上传回调参数
	}

	response, err := json.Marshal(policyToken)
	if err != nil {
		slog.Error("OSS json.Marshal fail", "error", err)
		return "", err
	}
	return string(response), nil
}

func (manager *AliyunOSSManager) Resign(objectName string) (string, error) {

	// 加载默认配置并设置凭证提供者和区域
	provider := sdk_credentials.NewStaticCredentialsProvider(
		manager.AccessKeyID,
		manager.AccessKeySecret,
	)

	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(provider).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 生成GetObject的预签名URL
	result, err := client.Presign(context.TODO(),
		&oss.GetObjectRequest{
			Bucket: oss.Ptr(bucketName),
			Key:    oss.Ptr(objectName),
		},
		oss.PresignExpires(15*time.Second), // 过期时间
	)
	if err != nil {
		slog.Error("failed to get object presign", "error", err)
		return "", err
	}
	if result != nil {
		slog.Info("Presign URL", "method", result.Method, "expiration", result.Expiration, "url", result.URL)
		return result.URL, nil
	}
	return "", errors.New("Presign Failed")
}
