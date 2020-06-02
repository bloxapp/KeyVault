package core

import (
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type EncryptableSeed struct {
	seed []byte
	encrypted map[string]interface{}
	encryptor types.Encryptor
}

func NewEncryptableSeed(seed []byte, encryptor types.Encryptor) *EncryptableSeed {
	return &EncryptableSeed{
		seed:      seed,
		encrypted: nil,
		encryptor: encryptor,
	}
}

func NewEncryptableSeedImmediatleyEncrypted(seed []byte, encryptor types.Encryptor, password []byte) (*EncryptableSeed,error) {
	ret := New(seed,encryptor)
	err := ret.Encrypt(password)
	if err != nil {
		return nil,err
	}
	return ret,nil
}

func (seed *EncryptableSeed) Seed() []byte {
	return seed.seed
}

func (seed *EncryptableSeed) IsEncrypted() bool {
	return seed.seed == nil
}

func (seed *EncryptableSeed) Encrypt(password []byte) error {
	ret,error :=  seed.encryptor.Encrypt(seed.seed,password)
	if error != nil {
		return error
	}

	seed.seed = nil
	seed.encrypted = ret
	return nil
}

func (seed *EncryptableSeed) Decrypt(password []byte) error {
	ret,error := seed.encryptor.Decrypt(seed.encrypted,password)
	if error != nil {
		return error
	}
	seed.seed = ret
	seed.encrypted = nil
	return nil
}
