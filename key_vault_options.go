package eth2keymanager

import (
	encryptor2 "github.com/bloxapp/eth2-key-manager/encryptor"
)

type KeyVaultOptions struct {
	encryptor encryptor2.Encryptor
	password  []byte
	storage   interface{} // a generic interface as there are a few core storage interfaces (storage, slashing storage and so on)
	seed      []byte
}

func (options *KeyVaultOptions) SetEncryptor(encryptor encryptor2.Encryptor) *KeyVaultOptions {
	options.encryptor = encryptor
	return options
}

func (options *KeyVaultOptions) SetStorage(storage interface{}) *KeyVaultOptions {
	options.storage = storage
	return options
}

func (options *KeyVaultOptions) SetPassword(password string) *KeyVaultOptions {
	options.password = []byte(password)
	return options
}

func (options *KeyVaultOptions) SetSeed(seed []byte) *KeyVaultOptions {
	options.seed = seed
	return options
}
