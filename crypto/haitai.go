package crypto

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"

	"github.com/AISHU-Technology/kweaver-go-lib/logger"
	"github.com/AISHU-Technology/kweaver-go-lib/rest"
)

const (
	HAITAI_VERSION string = "V2.0.2"
)

type haitaiCipher struct {
	key     string
	host    string
	realm   string
	dataKey string
}

func NewHaitaiCipher(key string, host string, realm string, dataKey string) Cipher {
	ci := &haitaiCipher{
		key:     key,
		host:    host,
		realm:   realm,
		dataKey: dataKey,
	}
	return ci
}

// hmac_sha256摘要k
func (ci haitaiCipher) hmacSha256(data []byte) string {
	key, _ := hex.DecodeString(ci.key)
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func (ci haitaiCipher) Encrypt(encryptedData string) string {
	logger.Fatalf("haitaiCipher Eecrypt is Not implemented Yet.")
	return ""
}

// 解密
func (ci haitaiCipher) Decrypt(encryptedData string) string {
	body := map[string]string{
		"data":    encryptedData,
		"dataKey": ci.dataKey,
	}
	bodyBytes, _ := sonic.Marshal(body)

	httpUrl := fmt.Sprintf("%s/ded-service/api/decrypt", ci.host)

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Digest algo=SHA256 realm=%s", ci.realm),
		"hmac":          ci.hmacSha256(bodyBytes),
		"timestamp":     strconv.FormatInt(time.Now().UnixNano()/1e6, 10),
		"version":       HAITAI_VERSION,
	}

	respCode, respData, err := rest.NewHTTPClient().Post(context.Background(), httpUrl, headers, body)
	if err != nil {
		logger.Fatalf("get request method failed: %v", err)
		return ""
	}
	if respCode != 200 {
		logger.Fatalf("get request method failed, httpCode: %v", respCode)
		return ""
	}

	resp := respData.(map[string]interface{})
	decryptedData := resp["data"].(string)
	return decryptedData
}

// 签名
func (ci haitaiCipher) Signature(signContent string) string {
	body := map[string]string{
		"data": signContent,
	}
	bodyBytes, _ := sonic.Marshal(body)

	httpUrl := fmt.Sprintf("%s/ded-service/api/sign", ci.host)

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Digest algo=SHA256 realm=%s", ci.realm),
		"hmac":          ci.hmacSha256(bodyBytes),
		"timestamp":     strconv.FormatInt(time.Now().UnixNano()/1e6, 10),
		"version":       HAITAI_VERSION,
	}

	respCode, respData, err := rest.NewHTTPClient().Post(context.Background(), httpUrl, headers, body)
	if err != nil {
		logger.Fatalf("get request method failed: %v", err)
		return ""
	}
	if respCode != 200 {
		logger.Fatalf("get request method failed, httpCode: %v", respCode)
		return ""
	}

	resp := respData.(map[string]interface{})
	signature := resp["data"].(string)
	return signature
}
