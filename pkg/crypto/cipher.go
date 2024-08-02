package crypto

import "os"

type Cipher interface {
	Encipher(data []byte) ([]byte, error)
	Decipher(data []byte) ([]byte, error)
}

// Cipher that performs no encryption whatsover
type CipherNone struct{}

func (cipher CipherNone) Encipher(data []byte) ([]byte, error) {
	return data, nil
}

func (cipher CipherNone) Decipher(data []byte) ([]byte, error) {
	return data, nil
}

func EncipherFile(cipher Cipher, filename string, data []byte) error {
	encrypted, err := cipher.Encipher(data)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, encrypted, 0600)
}

func DecipherFile(cipher Cipher, filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return cipher.Decipher(data)
}
