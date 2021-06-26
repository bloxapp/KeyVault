package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func _byteArray(input string) []byte {
	res, _ := hex.DecodeString(input)
	return res
}

func _bigIntFromSkHex(input string) *big.Int {
	res, _ := new(big.Int).SetString(input, 16)
	return res
}

func TestMarshalingHDKey(t *testing.T) {
	if err := InitBLS(); err != nil {
		os.Exit(1)
	}

	tests := []struct {
		name string
		seed []byte
		path string
		err  error
	}{
		{
			name: "validation account derivation",
			seed: _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path: "/0/0/0", // after basePath
			err:  nil,
		},
		{
			name: "withdrawal account derivation",
			seed: _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path: "/0/0", // after basePath
			err:  nil,
		},
		{
			name: "Base account derivation (base path only)",
			seed: _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path: "", // after basePath
			err:  errors.New("invalid relative path. Example: /1/2/3"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//storage := storage(test.seed, nil)
			//err := storage.SecurelySavePortfolioSeed(test.seed)
			//require.NoError(t, err)

			// create the privKey
			key, err := MasterKeyFromSeed(test.seed, PraterNetwork)
			require.NoError(t, err)

			hdKey, err := key.Derive(test.path, false)
			if test.err != nil {
				require.EqualError(t, test.err, err.Error())
				return
			} else {
				require.NoError(t, err)
			}

			// marshal and unmarshal
			byts, err := json.Marshal(hdKey)
			require.NoError(t, err)

			newKey := &HDKey{}
			err = json.Unmarshal(byts, newKey)
			require.NoError(t, err)

			// match
			require.Equal(t, hdKey.Path(), newKey.Path())
			require.Equal(t, hdKey.id.String(), newKey.id.String())
			require.Equal(t, hdKey.privKey, newKey.privKey)
			require.Equal(t, hdKey.PublicKey(), newKey.PublicKey())
		})
	}
}

func TestDerivableKeyRelativePathDerivation(t *testing.T) {
	if err := InitBLS(); err != nil {
		os.Exit(1)
	}

	tests := []struct {
		name        string
		seed        []byte
		path        string
		err         error
		expectedKey *big.Int
	}{
		{
			name:        "validation account 0 derivation",
			seed:        _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path:        "/0/0/0", // after basePath
			err:         nil,
			expectedKey: _bigIntFromSkHex("95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf"),
		},
		{
			name:        "validation account 1 derivation",
			seed:        _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path:        "/1/0/0", // after basePath
			err:         nil,
			expectedKey: _bigIntFromSkHex("b41df3c322a6fd305fc9425df52501f7f8067dbba551466d82d506c83c6ab287580202aa1a3449f54b9bc464a04b70e6"),
		},
		{
			name:        "validation account 2 derivation",
			seed:        _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path:        "/2/0/0", // after basePath
			err:         nil,
			expectedKey: _bigIntFromSkHex("9415b51f7996d6872f32c9bf7c259fad10e211d6097ff52ae99520a0ab3b916b3570073abbb83fa87da66936d351010d"),
		},
		{
			name:        "validation account 3 derivation",
			seed:        _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path:        "/3/0/0", // after basePath
			err:         nil,
			expectedKey: _bigIntFromSkHex("80b42ed53fe82598d055c2723bce9b1dde249d0497291856ef77fc75b094c60aca9dcc648e414dc9db41f8b8dc2f13e4"),
		},
		{
			name:        "withdrawal account 0 derivation",
			seed:        _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path:        "/0/0", // after basePath
			err:         nil,
			expectedKey: _bigIntFromSkHex("a0b9324da8a8a696c53950e984de25b299c123d17bab972eca1ac2c674964c9f817047bc6048ef0705d7ec6fae6d5da6"),
		},
		{
			name:        "withdrawal account 1 derivation",
			seed:        _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path:        "/1/0", // after basePath
			err:         nil,
			expectedKey: _bigIntFromSkHex("858e30df33bfdd613234abc9359ccd924f4807f1ba21de328d361e72f8c9ca94c9b7c225536405141df8239db87bd510"),
		},
		{
			name:        "withdrawal account 2 derivation",
			seed:        _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path:        "/2/0", // after basePath
			err:         nil,
			expectedKey: _bigIntFromSkHex("85586894abb77e41ba5dc3cfa2a7506c7584d024f028501da1e766792bcf6cd79ae17ff68eee84315eba9a2a8e7f89fe"),
		},
		{
			name:        "withdrawal account 3 derivation",
			seed:        _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path:        "/3/0", // after basePath
			err:         nil,
			expectedKey: _bigIntFromSkHex("afb22992f52aaf46c461ad1013e88c2c3ca8656c58170a9d08aaaeb9eac404fba839b313150f8f4b2f9fe23e64119c1f"),
		},
		{
			name:        "Base account derivation (big index)",
			seed:        _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path:        "/100/0", // after basePath
			err:         nil,
			expectedKey: _bigIntFromSkHex("aaa63a09aa2c0ce6a2a29940df8981eeefac0ea193bf90f2e06edd41356054f2907bc2e1eb5aaa4097361841914cd274"),
		},
		{
			name:        "bad path",
			seed:        _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path:        "0/0", // after basePath
			err:         errors.New("invalid relative path. Example: /1/2/3"),
			expectedKey: nil,
		},
		{
			name:        "too large of an index, bad path",
			seed:        _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path:        "/1000/0", // after basePath
			err:         errors.New("invalid relative path. Example: /1/2/3"),
			expectedKey: nil,
		},
		{
			name:        "not a relative path",
			seed:        _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			path:        "m/0/0", // after basePath
			err:         errors.New("invalid relative path. Example: /1/2/3"),
			expectedKey: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key, err := MasterKeyFromSeed(test.seed, MainNetwork)
			require.NoError(t, err)

			hdKey, err := key.Derive(test.path, false)
			if err != nil {
				if test.err != nil {
					require.Equal(t, test.err.Error(), err.Error())
				} else {
					t.Error(err)
				}
				return
			} else {
				require.NoError(t, test.err)
			}

			require.Equal(t, MainNetwork.FullPath(test.path), hdKey.Path())

			expectedPk := &bls.PublicKey{}
			require.NoError(t, expectedPk.Deserialize(test.expectedKey.Bytes()))

			require.NoError(t, err)
			require.Equal(t, expectedPk.Serialize(), hdKey.PublicKey().Serialize(), fmt.Sprintf("expected %s", hex.EncodeToString(hdKey.PublicKey().Serialize())))
		})
	}
}
