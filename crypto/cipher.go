package crypto

//go:generate mockgen -package mock -source ./cipher.go -destination ./mock/mock_cipher.go

type Cipher interface {
	Decrypt(encryptedData string) string
	Signature(signContent string) string
	Encrypt(data string) string
}
