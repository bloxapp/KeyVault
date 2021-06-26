package core

import (
	"encoding/hex"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	"regexp"

	"github.com/google/uuid"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/pkg/errors"
	util "github.com/wealdtech/go-eth2-util"
)

// EIP2334 paths.
const (
	// BaseEIP2334Path is the base EIP2334 path.
	BaseEIP2334Path = "m/12381/3600"
)

// MasterDerivableKey is responsible for managing privKey derivation, signing and  encryption.
// follows EIP 2333,2334
// MasterDerivableKey is not intended to be used a signing key, just as a medium for managing keys
type MasterDerivableKey struct {
	seed       []byte
	privateKey []byte
	network    Network
}

// MasterKeyFromSeed is the constructor of MasterDerivableKey.
// Base privKey is m / purpose / coin_type / as EIP 2334 defines
func MasterKeyFromSeed(seed []byte, network Network) (*MasterDerivableKey, error) {
	if seed == nil || len(seed) == 0 {
		return nil, errors.New("seed can't be nil or length 0")
	}
	return &MasterDerivableKey{
		seed:       seed,
		privateKey: nil,
		network:    network,
	}, nil
}

func MasterKeyFromPrivateKey(privateKey []byte, network Network) (*MasterDerivableKey, error) {
	if len(privateKey) == 0 {
		return nil, errors.New("private key is required")
	}
	return &MasterDerivableKey{
		seed:       nil,
		privateKey: privateKey,
		network:    network,
	}, nil
}

// Derive derives a HD key based on the given relative path.
func (master *MasterDerivableKey) Derive(relativePath string, seedless bool) (*HDKey, error) {
	if !validateRelativePath(relativePath) {
		return nil, errors.New("invalid relative path. Example: /1/2/3")
	}

	path := master.network.FullPath(relativePath)
	var key *e2types.BLSPrivateKey
	var err error

	if seedless == true {
		key, err = e2types.BLSPrivateKeyFromBytes(master.privateKey)
		if err != nil {
			return nil, err
		}
	} else {
		key, err = util.PrivateKeyFromSeedAndPath(master.seed, path) // TODO - needs to be refactored to remove wealdetch dependency
		if err != nil {
			return nil, err
		}
	}

	sk := &bls.SecretKey{}
	if err := sk.SetHexString(hex.EncodeToString(key.Marshal())); err != nil {
		return nil, err
	}

	return &HDKey{
		id:      uuid.New(),
		privKey: sk,
		path:    path,
	}, nil
}

func validateRelativePath(relativePath string) bool {
	match, _ := regexp.MatchString(`^(\/(\d\d?\d?))+$`, relativePath)
	return match
}
