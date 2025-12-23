package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strings"

	"github.com/AISHU-Technology/kweaver-go-lib/logger"
)

const (
	DEFAULT_CIPHER_MODE = "ECB"
)

type aesCipher struct {
	key        string
	cipherMode string
}

func NewAESCipher(key string) Cipher {
	ci := &aesCipher{
		key:        key,
		cipherMode: DEFAULT_CIPHER_MODE,
	}
	return ci
}

func (ci aesCipher) Decrypt(encryptedData string) string {
	switch ci.cipherMode {
	case "CBC":
		return ci.decryptCBC(encryptedData)
	case "ECB":
		return ci.decryptECB(encryptedData)
	default:
		logger.Fatalf("Invalid AES Cipher Mode: %s", ci.cipherMode)
		return ""
	}
}

func (ci aesCipher) Encrypt(encryptedData string) string {
	switch ci.cipherMode {
	case "ECB":
		return ci.encryptECB(encryptedData)
	default:
		logger.Fatalf("Invalid AES Cipher Mode: %s", ci.cipherMode)
		return ""
	}
}

// CBC方式解密
func (ci aesCipher) decryptCBC(encryptedData string) string {
	encrypted, _ := base64.StdEncoding.DecodeString(encryptedData)
	key := []byte(ci.key)
	// 创建实例
	block, _ := aes.NewCipher(key)
	//获取块的大小
	blockSize := block.BlockSize()
	//使用cbc
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	//初始化揭秘数据接收切片
	decrypted := make([]byte, len(encrypted))
	//执行解密
	blockMode.CryptBlocks(decrypted, encrypted)
	//去除填充
	decryptedData := ci.pkcs5UnPadding(decrypted)
	return string(decryptedData)
}

// pkcs5方式解除填充
func (ci aesCipher) pkcs5UnPadding(decryptedData []byte) []byte {
	length := len(decryptedData)
	unpadding := int(decryptedData[length-1])
	return decryptedData[:(length - unpadding)]
}

// ECB方式加密
func (ci aesCipher) encryptECB(data string) string {

	cipherText := []byte(data)
	key := []byte(ci.key)
	//创建实例
	block, _ := aes.NewCipher(ci.generateKey(key))
	//获取块的大小
	blockSize := block.BlockSize()

	length := (len(cipherText) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, cipherText)

	encrypted := make([]byte, len(plain))
	//分组分块加密
	for bs, be := 0, blockSize; bs <= len(cipherText); bs, be = bs+blockSize, be+blockSize {
		block.Encrypt(encrypted[bs:be], plain[bs:be])
	}

	b := base64.StdEncoding.EncodeToString(encrypted)
	return b
}

// ECB方式解密
func (ci aesCipher) decryptECB(encryptedData string) string {
	encrypted, _ := base64.StdEncoding.DecodeString(encryptedData)
	key := []byte(ci.key)
	// 创建实例
	block, _ := aes.NewCipher(ci.generateKey(key))
	//获取块的大小
	blockSize := block.BlockSize()
	//初始化揭秘数据接收切片
	decrypted := make([]byte, len(encrypted))
	//分组分块解密
	for bs, be := 0, blockSize; bs < len(encrypted); bs, be = bs+blockSize, be+blockSize {
		block.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}

	trim := 0
	if len(decrypted) > 0 {
		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
	}

	decryptedData := strings.TrimRight(string(decrypted[:trim]), "\x00")
	return decryptedData
}

func (ci aesCipher) generateKey(key []byte) (genKey []byte) {
	genKey = make([]byte, 16)
	copy(genKey, key)
	for i := 16; i < len(key); {
		for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
			genKey[j] ^= key[i]
		}
	}
	return genKey
}

func (ci aesCipher) Signature(signContent string) string {
	logger.Fatalf("aesCipher Signature is Not implemented Yet.")
	return ""
}
