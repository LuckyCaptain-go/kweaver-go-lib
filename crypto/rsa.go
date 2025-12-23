package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strings"

	"github.com/AISHU-Technology/kweaver-go-lib/logger"
)

const (
	PEM_BEGIN = "-----BEGIN RSA PRIVATE KEY-----\n"
	PEM_END   = "\n-----END RSA PRIVATE KEY-----"
)

type rsaCipher struct {
	privateKey string
	publicKey  string
}

func NewRSACipher(privateKey string, publicKey string) Cipher {
	ci := &rsaCipher{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
	return ci
}

func (ci rsaCipher) Decrypt(encryptedData string) string {
	logger.Fatalf("rsaCipher Decrypt is Not implemented Yet.")
	return ""
}

func (ci rsaCipher) Encrypt(encryptedData string) string {
	logger.Fatalf("rsaCipher Eecrypt is Not implemented Yet.")
	return ""
}

// RSA方式签名
func (ci rsaCipher) Signature(signContent string) string {
	shaNew := sha256.New()
	shaNew.Write([]byte(signContent))
	hashed := shaNew.Sum(nil)
	priKey, err := ci.parsePrivateKey(ci.privateKey)
	if err != nil {
		logger.Fatalf("%v", err)
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, hashed)
	if err != nil {
		logger.Fatalf("%v", err)
	}
	encodedSign := base64.StdEncoding.EncodeToString(signature)
	return encodedSign
}

func (ci rsaCipher) parsePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	privateKey = ci.formatPrivateKey(privateKey)
	// 2、解码私钥字节，生成加密对象
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return nil, errors.New("私钥信息错误！")
	}
	// 3、解析DER编码的私钥，生成私钥对象
	priKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return priKey, nil
}

func (ci rsaCipher) formatPrivateKey(privateKey string) string {
	if !strings.HasPrefix(privateKey, PEM_BEGIN) {
		privateKey = PEM_BEGIN + privateKey
	}
	if !strings.HasSuffix(privateKey, PEM_END) {
		privateKey = privateKey + PEM_END
	}
	return privateKey
}
