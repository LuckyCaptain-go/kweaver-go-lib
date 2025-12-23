package crypto

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	KEY   = "key"
	ODATA = "value"
	EDATA = "bDLlZGAVKoUlmFM9FEPzrQ=="
)

func TestAesDecrypt(t *testing.T) {
	Convey("test decrypt\n", t, func() {
		aesCipher := NewAESCipher(KEY)

		Convey("ECB mode test \n", func() {
			a := aesCipher.Decrypt(EDATA)
			So(a, ShouldEqual, ODATA)
		})
	})
}

func TestAesEncrypt(t *testing.T) {
	Convey("test decrypt\n", t, func() {
		aesCipher := NewAESCipher(KEY)

		Convey("ECB mode test \n", func() {
			a := aesCipher.Encrypt(ODATA)
			So(a, ShouldEqual, EDATA)
		})
	})
}
