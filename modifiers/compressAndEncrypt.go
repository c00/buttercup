package modifiers

import (
	"fmt"
	"io"
	"os"

	"filippo.io/age"
	"github.com/klauspost/compress/zstd"
)

func CompressAndEncryptFile(input, output, passphrase string) error {
	reader, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("cannot open input path: %w", err)
	}
	defer reader.Close()

	writer, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("cannot open output path: %w", err)
	}
	defer writer.Close()

	err = CompressAndEncrypt(reader, writer, passphrase)
	if err != nil {
		return fmt.Errorf("cannot encrypt: %w", err)
	}
	return nil
}

// todo write tests.
func CompressAndEncrypt(input io.Reader, output io.Writer, password string) error {
	recipient, err := age.NewScryptRecipient(password)
	if err != nil {
		return err
	}
	encryptedWriter, err := age.Encrypt(output, recipient)
	if err != nil {
		return err
	}
	defer encryptedWriter.Close()

	compressionWriter, err := zstd.NewWriter(encryptedWriter)
	if err != nil {
		return err
	}
	defer compressionWriter.Close()

	_, err = io.Copy(compressionWriter, input)
	if err != nil {
		return err
	}

	return nil
}
