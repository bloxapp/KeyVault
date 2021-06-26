package wallets

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/bloxapp/eth2-key-manager/stores/dummy"
)

func _byteArray(input string) []byte {
	res, _ := hex.DecodeString(input)
	return res
}

func storage() core.Storage {
	return &dummy.Storage{}
}

func TestAccountMarshaling(t *testing.T) {
	tests := []struct {
		id       uuid.UUID
		testName string
		//accountType core.AccountType
		parentWalletID uuid.UUID
		name           string
		seed           []byte
		accountIndex   string
	}{
		{
			testName: "simple account",
			id:       uuid.New(),
			//accountType:core.ValidatorAccount,
			parentWalletID: uuid.New(),
			name:           "account1",
			seed:           _byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
			accountIndex:   "0",
		},
	}

	core.InitBLS()

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// setup storage
			storage := storage()

			// create key and account
			masterKey, err := core.MasterKeyFromSeed(test.seed, core.PraterNetwork)
			require.NoError(t, err)
			validationKey, err := masterKey.Derive(fmt.Sprintf("/%s/0/0", test.accountIndex), false)
			require.NoError(t, err)
			withdrawalKey, err := masterKey.Derive(fmt.Sprintf("/%s/0", test.accountIndex), false)
			require.NoError(t, err)
			a := &HDAccount{
				//accountType:test.accountType,
				name:             test.name,
				id:               test.id,
				validationKey:    validationKey,
				withdrawalPubKey: withdrawalKey.PublicKey().Serialize(),
				basePath:         fmt.Sprintf("/%s", test.accountIndex),
			}

			// marshal
			byts, err := json.Marshal(a)
			require.NoError(t, err)
			//unmarshal
			a1 := &HDAccount{context: &core.WalletContext{Storage: storage}}
			err = json.Unmarshal(byts, a1)
			require.NoError(t, err)

			require.Equal(t, a.id, a1.id)
			require.Equal(t, a.name, a1.name)
			require.Equal(t, a.validationKey.PublicKey().Serialize(), a1.validationKey.PublicKey().Serialize())
			require.Equal(t, a.withdrawalPubKey, a1.withdrawalPubKey)
			require.Equal(t, a.basePath, a1.basePath)
		})
	}
}
