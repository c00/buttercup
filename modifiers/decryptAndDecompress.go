package modifiers

import (
	"fmt"
	"io"
	"os"

	"filippo.io/age"
	"github.com/klauspost/compress/zstd"
)

func DecryptAndDecompressFile(inputPath, outputPath, passphrase string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("cannot open input file: %w", err)
	}

	reader, err := DecryptAndDecompress(file, passphrase)
	if err != nil {
		return fmt.Errorf("cannot get reader: %w", err)
	}
	defer reader.Close()

	writer, err := os.OpenFile(outputPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("cannot open output file: %w", err)
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("cannot write decrypt: %w", err)
	}

	return nil
}

// todo write tests.
func DecryptAndDecompress(input io.Reader, password string) (io.ReadCloser, error) {
	identity, err := age.NewScryptIdentity(password)
	if err != nil {
		return nil, err
	}
	decryptedReader, err := age.Decrypt(input, identity)
	if err != nil {
		return nil, err
	}

	decompressedReader, err := zstd.NewReader(decryptedReader)
	if err != nil {
		return nil, err
	}

	return decompressedReader.IOReadCloser(), nil
}
