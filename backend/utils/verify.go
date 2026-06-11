package utils

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ossPubKeyCache  = sync.Map{}
	ossPubKeyClient = &http.Client{
		Timeout: 5 * time.Second,
	}
	fetchOSSPublicKey = defaultFetchOSSPublicKey
)

// VerifyOSS 对 OSS 回调请求进行验签
func VerifyOSS(r *http.Request) (bool, error) {
	signatureBase64 := strings.TrimSpace(r.Header.Get("Authorization"))
	if signatureBase64 == "" {
		return false, errors.New("missing authorization header")
	}

	pubKeyURLBase64 := strings.TrimSpace(r.Header.Get("x-oss-pub-key-url"))
	if pubKeyURLBase64 == "" {
		return false, errors.New("missing x-oss-pub-key-url header")
	}

	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false, fmt.Errorf("decode authorization failed: %w", err)
	}

	pubKeyURLBytes, err := base64.StdEncoding.DecodeString(pubKeyURLBase64)
	if err != nil {
		return false, fmt.Errorf("decode pub key url failed: %w", err)
	}

	pubKeyURL := string(pubKeyURLBytes)
	if !strings.HasPrefix(pubKeyURL, "http://gosspublic.alicdn.com/") &&
		!strings.HasPrefix(pubKeyURL, "https://gosspublic.alicdn.com/") {
		return false, errors.New("invalid oss public key url")
	}

	pubKey, err := fetchOSSPublicKey(pubKeyURL)
	if err != nil {
		return false, err
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false, fmt.Errorf("read request body failed: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	signStr, err := buildOSSCallbackSignString(r, body)
	if err != nil {
		return false, err
	}

	digest := md5.Sum([]byte(signStr))
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.MD5, digest[:], signature); err != nil {
		return false, fmt.Errorf("verify oss callback signature failed: %w", err)
	}

	return true, nil
}

func buildOSSCallbackSignString(r *http.Request, body []byte) (string, error) {
	decodedPath, err := url.PathUnescape(r.URL.EscapedPath())
	if err != nil {
		return "", fmt.Errorf("unescape callback path failed: %w", err)
	}

	var builder strings.Builder
	builder.WriteString(decodedPath)
	if r.URL.RawQuery != "" {
		builder.WriteByte('?')
		builder.WriteString(r.URL.RawQuery)
	}
	builder.WriteByte('\n')
	builder.Write(body)

	return builder.String(), nil
}

func defaultFetchOSSPublicKey(pubKeyURL string) (*rsa.PublicKey, error) {
	if cached, ok := ossPubKeyCache.Load(pubKeyURL); ok {
		pubKey, ok := cached.(*rsa.PublicKey)
		if ok {
			return pubKey, nil
		}
	}

	resp, err := ossPubKeyClient.Get(pubKeyURL)
	if err != nil {
		return nil, fmt.Errorf("request oss public key failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request oss public key failed, status=%d", resp.StatusCode)
	}

	pubKeyPEM, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read oss public key failed: %w", err)
	}

	pubKey, err := parseOSSPublicKey(pubKeyPEM)
	if err != nil {
		return nil, err
	}

	ossPubKeyCache.Store(pubKeyURL, pubKey)
	return pubKey, nil
}

func parseOSSPublicKey(pubKeyPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubKeyPEM)
	if block == nil {
		return nil, errors.New("decode pem public key failed")
	}

	if pubKeyAny, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		pubKey, ok := pubKeyAny.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("oss public key is not rsa")
		}
		return pubKey, nil
	}

	if pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return pubKey, nil
	}

	if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("oss certificate public key is not rsa")
		}
		return pubKey, nil
	}

	return nil, errors.New("parse oss public key failed")
}
