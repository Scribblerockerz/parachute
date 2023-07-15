package archive

import (
	"io/ioutil"
	"log"

	openssl "github.com/Luzifer/go-openssl/v4"
)

func EncryptFile(sourcePath string, targetPath string, passphrase string) error {
	plainText, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	o := openssl.New()

	cipherText, err := o.EncryptBinaryBytes(passphrase, []byte(plainText), openssl.PBKDF2SHA256)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(targetPath, cipherText, 0777)
	if err != nil {
		return err
	}

	return nil
}

func DecryptFile(sourcePath string, targetPath string, passphrase string) error {
	cipherText, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		log.Fatal(err)
	}

	o := openssl.New()

	plainText, err := o.DecryptBinaryBytes(passphrase, []byte(cipherText), openssl.PBKDF2SHA256)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(targetPath, plainText, 0777)
	if err != nil {
		return err
	}

	return nil
}
