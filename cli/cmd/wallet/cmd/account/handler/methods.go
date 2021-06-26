package handler

import (
	"encoding/hex"
	cmd2 "github.com/bloxapp/eth2-key-manager/cli/cmd"
	"github.com/pkg/errors"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/spf13/cobra"

	eth2keymanager "github.com/bloxapp/eth2-key-manager"
	"github.com/bloxapp/eth2-key-manager/cli/cmd/wallet/cmd/account/flag"
	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/bloxapp/eth2-key-manager/stores/inmemory"
)

// CreateAccountFlagValues keeps all collected values for seed and seedless modes
type CreateAccountFlagValues struct {
	index            int
	seed             string
	seedBytes        []byte
	privateKey       []byte
	accumulate       bool
	responseType     flag.ResponseType
	highestSources   []uint64
	highestTargets   []uint64
	highestProposals []uint64
	network          core.Network
}

// CheckHighestValues Performs basic checks for account flags
func CheckHighestValues(accountFlagValues CreateAccountFlagValues) error {
	if accountFlagValues.accumulate {
		if len(accountFlagValues.highestSources) != (accountFlagValues.index + 1) {
			return errors.Errorf("highest sources length when the accumulate flag is true need to be equal to index")
		}
		if len(accountFlagValues.highestTargets) != (accountFlagValues.index + 1) {
			return errors.Errorf("highest targets length when the accumulate flag is true need to be index")
		}
		if len(accountFlagValues.highestProposals) != (accountFlagValues.index + 1) {
			return errors.Errorf("highest proposals length when the accumulate flag is true need to be index")
		}
	} else {
		if len(accountFlagValues.highestSources) != 1 {
			return errors.Errorf("highest sources length when the accumulate flag is false need to be 1")
		}
		if len(accountFlagValues.highestTargets) != 1 {
			return errors.Errorf("highest targets length when the accumulate flag is false need to be 1")
		}
		if len(accountFlagValues.highestProposals) != 1 {
			return errors.Errorf("highest proposals length when the accumulate flag is false need to be 1")
		}
	}
	return nil
}

// CollectAccountFlags returns collected flags for seed and seedless modes
func CollectAccountFlags(cmd *cobra.Command, seedless bool) (*CreateAccountFlagValues, error) {
	accountFlagValues := CreateAccountFlagValues{}

	// Get index flag.
	indexFlagValue, err := flag.GetIndexFlagValue(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the index flag value")
	}
	accountFlagValues.index = indexFlagValue

	// Seedless mode
	if seedless == true {
		// Private key
		privateKeyValue, err := flag.GetPrivateKeyFlagValue(cmd)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve the private key flag value")
		}
		privateKeyBytes, err := hex.DecodeString(privateKeyValue)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert private key string to bytes")
		}
		accountFlagValues.privateKey = privateKeyBytes

		// Seed mode
	} else {
		// Seed flag
		seedFlagValue, err := flag.GetSeedFlagValue(cmd)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve the seed flag value")
		}
		accountFlagValues.seed = seedFlagValue

		// Get seed bytes
		seedBytes, err := hex.DecodeString(seedFlagValue)
		if err != nil {
			return nil, errors.Wrap(err, "failed to HEX decode seed")
		}
		accountFlagValues.seedBytes = seedBytes
	}

	// Get accumulate flag.
	accumulateFlagValue, err := flag.GetAccumulateFlagValue(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the accumulate flag value")
	}
	accountFlagValues.accumulate = accumulateFlagValue

	// Get response-type flag.
	responseType, err := flag.GetResponseTypeFlagValue(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the response type value")
	}
	accountFlagValues.responseType = responseType

	// Get minimals slashing data flag
	highestSources, err := flag.GetHighestSourceFlagValue(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the minimal slashing data value")
	}
	accountFlagValues.highestSources = highestSources

	highestTargets, err := flag.GetHighestTargetFlagValue(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the minimal slashing data value")
	}
	accountFlagValues.highestTargets = highestTargets

	highestProposals, err := flag.GetHighestProposalFlagValue(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the minimal slashing data value")
	}
	accountFlagValues.highestProposals = highestProposals

	// Check highest values
	highestValuesError := CheckHighestValues(accountFlagValues)
	if highestValuesError != nil {
		return nil, highestValuesError
	}

	// Get network value
	network, err := cmd2.GetNetworkFlagValue(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to network")
	}
	accountFlagValues.network = network

	return &accountFlagValues, nil
}

// GenerateOneAccount generates account by index using provided account flags
func GenerateOneAccount(wallet core.Wallet, store *inmemory.InMemStore, index int, accountFlags *CreateAccountFlagValues, seedless bool) error {

	var acc core.ValidatorAccount
	var err error

	if seedless == true {
		acc, err = wallet.CreateValidatorAccountFromPrivateKey(accountFlags.privateKey, &index)
		if err != nil {
			return errors.Wrap(err, "failed to create validator account")
		}
	} else {
		acc, err = wallet.CreateValidatorAccount(accountFlags.seedBytes, &index)
		if err != nil {
			return errors.Wrap(err, "failed to create validator account")
		}
	}

	highestIndex := index
	if seedless == true || accountFlags.accumulate != true {
		highestIndex = 0
	}

	// add minimal attestation protection data
	minimalAtt := &eth.AttestationData{
		Source: &eth.Checkpoint{Epoch: uint64(accountFlags.highestSources[highestIndex])},
		Target: &eth.Checkpoint{Epoch: uint64(accountFlags.highestTargets[highestIndex])},
	}
	if err := store.SaveHighestAttestation(acc.ValidatorPublicKey(), minimalAtt); err != nil {
		return errors.Wrap(err, "failed to set validator minimal slashing protection")
	}

	// add minimal proposal protection data
	minimalProposal := &eth.BeaconBlock{
		Slot: uint64(accountFlags.highestProposals[highestIndex]),
	}
	if err := store.SaveHighestProposal(acc.ValidatorPublicKey(), minimalProposal); err != nil {
		return errors.Wrap(err, "failed to set validator minimal slashing protection")
	}
	return nil
}

// BuildAndPrintAccounts builds all accounts or one account depending of seedless flag
func (h *Account) BuildAndPrintAccounts(accountFlags *CreateAccountFlagValues, seedless bool) error {
	// Initialize store
	store := inmemory.NewInMemStore(accountFlags.network)
	options := &eth2keymanager.KeyVaultOptions{}
	options.SetStorage(store)

	// Create new key vault
	_, err := eth2keymanager.NewKeyVault(options)
	if err != nil {
		return errors.Wrap(err, "failed to create key vault")
	}

	wallet, err := store.OpenWallet()
	if err != nil {
		return errors.Wrap(err, "failed to open wallet")
	}

	if accountFlags.accumulate && seedless != true {
		for i := 0; i <= accountFlags.index; i++ {
			err := GenerateOneAccount(wallet, store, i, accountFlags, seedless)
			if err != nil {
				return err
			}
		}
	} else {
		err := GenerateOneAccount(wallet, store, accountFlags.index, accountFlags, seedless)
		if err != nil {
			return err
		}
	}

	if accountFlags.responseType == flag.StorageResponseType {
		// marshal storage
		bytes, err := store.MarshalJSON()
		if err != nil {
			return errors.Wrap(err, "failed to JSON marshal storage")
		}

		h.printer.Text(hex.EncodeToString(bytes))
		return nil
	}

	var accounts []map[string]string
	for _, a := range wallet.Accounts() {
		var withdrawalPubKey string
		if seedless == true {
			withdrawalPubKey = ""
		} else {
			withdrawalPubKey = hex.EncodeToString(a.WithdrawalPublicKey())
		}
		accObj := map[string]string{
			"id":               a.ID().String(),
			"name":             a.Name(),
			"validationPubKey": hex.EncodeToString(a.ValidatorPublicKey()),
			"withdrawalPubKey": withdrawalPubKey,
		}
		accounts = append(accounts, accObj)
	}

	if accountFlags.accumulate {
		err = h.printer.JSON(accounts)
	} else if len(accounts) > 0 {
		err = h.printer.JSON(accounts[0])
	}
	if err != nil {
		return errors.Wrap(err, "failed to print accounts JSON")
	}
	return nil
}
