package dpapi

import (
	"github.com/billgraziano/dpapi"
)

/*
 * Utilize the Windows DPAPI for system-level encryption
 */

type CipherDPAPI struct{}

func (cipher CipherDPAPI) Encipher(data []byte) ([]byte, error) {
	encrypted, err := dpapi.EncryptBytesMachineLocal(data)
	return encrypted, err
}

func (cipher CipherDPAPI) Decipher(data []byte) ([]byte, error) {
	decrypted, err := dpapi.DecryptBytes(data)
	return decrypted, err
}
