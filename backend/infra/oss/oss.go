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
	sdkcredentials "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/aliyun/credentials-go/credentials"
	"github.com/yzletter/go-postery/backend/conf"
)

var (
	region      string
	bucketName  string
	product     = "oss"
	callbackURL string
)

type Manager interface {
	Sign(dir string, uploadCallbackURL string) (string, error)
	Resign(objectName string) (string, error)
}

type PolicyToken struct {
	Policy           string `json:"policy"`
	SecurityToken    string `json:"security_token"`
	SignatureVersion string `json:"x_oss_signature_version"`
	Credential       string `json:"x_oss_credential"`
	Date             string `json:"x_oss_date"`
	Signature        string `json:"signature"`
	Host             string `json:"host"`
	Dir              string `json:"dir"`
	Callback         string `json:"callback"`
}

type CallbackParam struct {
	CallbackURL      string `json:"callbackUrl"`
	CallbackBody     string `json:"callbackBody"`
	CallbackBodyType string `json:"callbackBodyType"`
}

type AliyunOSSManager struct {
	AccessKeyID     string
	AccessKeySecret string
	Arn             string
}

func Init(config conf.OSSConfig) *AliyunOSSManager {
	region = config.Region
	if region == "" {
		region = "cn-hongkong"
	}
	bucketName = config.Bucket
	if bucketName == "" {
		bucketName = "go-postery"
	}
	callbackURL = config.CallbackURL
	if callbackURL == "" {
		callbackURL = "http://gopostery.top/api/v1/users/callback"
	}

	return &AliyunOSSManager{
		AccessKeyID:     config.AccessKeyID,
		AccessKeySecret: config.AccessKeySecret,
		Arn:             config.Arn,
	}
}

func (manager *AliyunOSSManager) Sign(dir string, uploadCallbackURL string) (string, error) {
	if uploadCallbackURL == "" {
		uploadCallbackURL = callbackURL
	}
	host := fmt.Sprintf("https://%s.oss-%s.aliyuncs.com", bucketName, region)

	config := new(credentials.Config).
		SetType("ram_role_arn").
		SetAccessKeyId(manager.AccessKeyID).
		SetAccessKeySecret(manager.AccessKeySecret).
		SetRoleArn(manager.Arn).
		SetRoleSessionName("Upload").
		SetPolicy("").
		SetRoleSessionExpiration(3600)

	provider, err := credentials.NewCredential(config)
	if err != nil {
		slog.Error("OSS NewCredential Failed", "error", err)
		return "", err
	}

	cred, err := provider.GetCredential()
	if err != nil {
		slog.Error("OSS GetCredential Failed", "error", err)
		return "", err
	}

	utcTime := time.Now().UTC()
	date := utcTime.Format("20060102")
	expiration := utcTime.Add(time.Hour)
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

	policy, err := json.Marshal(policyMap)
	if err != nil {
		slog.Error("OSS Marshal Policy Failed", "error", err)
		return "", err
	}

	stringToSign := base64.StdEncoding.EncodeToString(policy)

	hmacHash := func() hash.Hash { return sha256.New() }
	signingKey := "aliyun_v4" + *cred.AccessKeySecret
	h1 := hmac.New(hmacHash, []byte(signingKey))
	_, _ = io.WriteString(h1, date)
	h1Key := h1.Sum(nil)

	h2 := hmac.New(hmacHash, h1Key)
	_, _ = io.WriteString(h2, region)
	h2Key := h2.Sum(nil)

	h3 := hmac.New(hmacHash, h2Key)
	_, _ = io.WriteString(h3, product)
	h3Key := h3.Sum(nil)

	h4 := hmac.New(hmacHash, h3Key)
	_, _ = io.WriteString(h4, "aliyun_v4_request")
	h4Key := h4.Sum(nil)

	h := hmac.New(hmacHash, h4Key)
	_, _ = io.WriteString(h, stringToSign)
	signature := hex.EncodeToString(h.Sum(nil))

	callbackParam := CallbackParam{
		CallbackURL:      uploadCallbackURL,
		CallbackBody:     `{"bucket":${bucket},"object":${object},"size":${size}}`,
		CallbackBodyType: "application/json",
	}
	callbackPayload, err := json.Marshal(callbackParam)
	if err != nil {
		slog.Error("OSS Marshal Callback Failed", "error", err)
		return "", err
	}
	callbackBase64 := base64.StdEncoding.EncodeToString(callbackPayload)

	policyToken := PolicyToken{
		Policy:           stringToSign,
		SecurityToken:    *cred.SecurityToken,
		SignatureVersion: "OSS4-HMAC-SHA256",
		Credential:       fmt.Sprintf("%v/%v/%v/%v/aliyun_v4_request", *cred.AccessKeyId, date, region, product),
		Date:             utcTime.Format("20060102T150405Z"),
		Signature:        signature,
		Host:             host,
		Dir:              dir,
		Callback:         callbackBase64,
	}

	response, err := json.Marshal(policyToken)
	if err != nil {
		slog.Error("OSS Marshal Response Failed", "error", err)
		return "", err
	}
	return string(response), nil
}

func (manager *AliyunOSSManager) Resign(objectName string) (string, error) {
	provider := sdkcredentials.NewStaticCredentialsProvider(
		manager.AccessKeyID,
		manager.AccessKeySecret,
	)

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(provider).
		WithRegion(region)

	client := oss.NewClient(cfg)
	result, err := client.Presign(context.TODO(),
		&oss.GetObjectRequest{
			Bucket: oss.Ptr(bucketName),
			Key:    oss.Ptr(objectName),
		},
		oss.PresignExpires(30*time.Minute),
	)
	if err != nil {
		slog.Error("Get OSS Object Presign Failed", "error", err)
		return "", err
	}
	if result != nil {
		slog.Info("Presign URL", "method", result.Method, "expiration", result.Expiration, "url", result.URL)
		return result.URL, nil
	}
	return "", errors.New("presign failed")
}
