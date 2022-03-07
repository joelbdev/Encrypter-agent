package Encrypt

import (
	"log"
	"os"
	"path"
	"path/filepath"
)

type Encrypter struct {
	FileExtension string
}

func (c *Encrypter) Encrypt(startDirectory string) error {
	start := path.Join(startDirectory)

	err := filepath.Walk(start, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}
		if info.Size() != 0 {
			err := EncryptFile(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Println(err)
	}

	//fileInfo, err := os.Stat(file)
	//if fileInfo.IsDir() {
	//cd in and go to files
	return nil
}
