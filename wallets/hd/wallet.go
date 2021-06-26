package hd

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/bloxapp/eth2-key-manager/wallets"
)

// Default values according to https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2334.md
const (
	BaseAccountPath   = "/%d"
	WithdrawalKeyPath = BaseAccountPath + "/0"
	ValidatorKeyPath  = WithdrawalKeyPath + "/0"
)

// Predefined errors
var (
	// ErrAccountNotFound is the error when account not found
	ErrAccountNotFound = errors.New("account not found")
)

// Wallet represents hierarchical deterministic wallet
type Wallet struct {
	id          uuid.UUID
	walletType  core.WalletType
	indexMapper map[string]uuid.UUID
	context     *core.WalletContext
}

// NewWallet is the constructor of Wallet
func NewWallet(context *core.WalletContext) *Wallet {
	return &Wallet{
		id:          uuid.New(),
		walletType:  core.HDWallet,
		indexMapper: make(map[string]uuid.UUID),
		context:     context,
	}
}

// ID provides the ID for the wallet.
func (wallet *Wallet) ID() uuid.UUID {
	return wallet.id
}

// Type provides the type of the wallet.
func (wallet *Wallet) Type() core.WalletType {
	return wallet.walletType
}

// GetNextAccountIndex provides next index to create account at.
func (wallet *Wallet) GetNextAccountIndex() int {
	if len(wallet.indexMapper) == 0 {
		return 0
	}
	accounts := wallet.Accounts()
	index, _ := strconv.ParseInt(accounts[0].BasePath()[1:], 0, 64)
	return int(index) + 1
}

// BuildValidatorAccount using pointer and constructed key, using seedless or seed modes
func (wallet *Wallet) BuildValidatorAccount(indexPointer *int, key *core.MasterDerivableKey, seedless bool) (*wallets.HDAccount, error) {
	// Resolve index to create account at
	var index int
	if indexPointer != nil {
		index = *indexPointer
	} else {
		index = wallet.GetNextAccountIndex()
	}
	name := fmt.Sprintf("account-%d", index)

	baseAccountPath := fmt.Sprintf(BaseAccountPath, index)

	// Create validator key
	validatorPath := fmt.Sprintf(ValidatorKeyPath, index)
	validatorKey, err := key.Derive(validatorPath, seedless)
	if err != nil {
		return nil, err
	}

	// Create withdrawal key
	withdrawalPath := fmt.Sprintf(WithdrawalKeyPath, index)
	withdrawalKey, err := key.Derive(withdrawalPath, seedless)
	if err != nil {
		return nil, err
	}

	// Create ret account
	ret := wallets.NewValidatorAccount(
		name,
		validatorKey,
		withdrawalKey.PublicKey().Serialize(),
		baseAccountPath,
		wallet.context,
	)

	validatorPublicKey := hex.EncodeToString(ret.ValidatorPublicKey())

	// Register new wallet and save portfolio
	reset := func() {
		delete(wallet.indexMapper, validatorPublicKey)
	}
	wallet.indexMapper[validatorPublicKey] = ret.ID()

	// Store account
	if err = wallet.context.Storage.SaveAccount(ret); err != nil {
		reset()
		return nil, err
	}

	// Store wallet
	err = wallet.context.Storage.SaveWallet(wallet)
	if err != nil {
		reset()
		return nil, err
	}

	return ret, nil
}

// CreateValidatorAccountFromPrivateKey creates account having only private key
func (wallet *Wallet) CreateValidatorAccountFromPrivateKey(privateKey []byte, indexPointer *int) (core.ValidatorAccount, error) {
	// Create the master key based on the private key and network.
	key, err := core.MasterKeyFromPrivateKey(privateKey, wallet.context.Storage.Network())
	if err != nil {
		return nil, err
	}

	// Build account
	account, err := wallet.BuildValidatorAccount(indexPointer, key, true)
	if err != nil {
		return nil, err
	}
	return account, nil
}

// CreateValidatorAccount creates a new validation (validator) key pair in the wallet.
func (wallet *Wallet) CreateValidatorAccount(seed []byte, indexPointer *int) (core.ValidatorAccount, error) {
	// Create the master key based on the seed and network.
	key, err := core.MasterKeyFromSeed(seed, wallet.context.Storage.Network())
	if err != nil {
		return nil, err
	}

	// Build account
	account, err := wallet.BuildValidatorAccount(indexPointer, key, false)
	if err != nil {
		return nil, err
	}
	return account, nil
}

// AddValidatorAccount returns error
func (wallet *Wallet) AddValidatorAccount(account core.ValidatorAccount) error {
	validatorPublicKey := hex.EncodeToString(account.ValidatorPublicKey())
	wallet.indexMapper[validatorPublicKey] = account.ID()

	// Store account
	if err := wallet.context.Storage.SaveAccount(account); err != nil {
		return err
	}

	// Store wallet
	err := wallet.context.Storage.SaveWallet(wallet)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAccountByPublicKey deletes account by the given public key
func (wallet *Wallet) DeleteAccountByPublicKey(pubKey string) error {
	account, err := wallet.AccountByPublicKey(pubKey)
	if err != nil {
		return errors.Wrap(err, "failed to get account by public key")
	}

	if err := wallet.context.Storage.DeleteAccount(account.ID()); err != nil {
		return errors.Wrap(err, "failed to delete account from store")
	}
	delete(wallet.indexMapper, pubKey)

	if err := wallet.context.Storage.SaveWallet(wallet); err != nil {
		return errors.Wrap(err, "failed to save wallet")
	}
	return nil
}

// Accounts provides all accounts in the wallet.
func (wallet *Wallet) Accounts() []core.ValidatorAccount {
	accounts := make([]core.ValidatorAccount, 0)
	for pubKey := range wallet.indexMapper {
		id := wallet.indexMapper[pubKey]
		account, err := wallet.AccountByID(id)
		if err != nil {
			continue
		}
		accounts = append(accounts, account)
	}
	sort.Slice(accounts, func(i, j int) bool {
		a, _ := strconv.ParseInt(accounts[i].BasePath()[1:], 0, 64)
		b, _ := strconv.ParseInt(accounts[j].BasePath()[1:], 0, 64)
		return a > b
	})
	return accounts
}

// AccountByID provides a nd account from the wallet given its ID.
// This will error if the account is not found.
func (wallet *Wallet) AccountByID(id uuid.UUID) (core.ValidatorAccount, error) {
	ret, err := wallet.context.Storage.OpenAccount(id)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return nil, ErrAccountNotFound
	}

	ret.SetContext(wallet.context)
	return ret, nil
}

// SetContext is the context setter
func (wallet *Wallet) SetContext(ctx *core.WalletContext) {
	wallet.context = ctx
}

// AccountByPublicKey provides a nd account from the wallet given its public key.
// This will error if the account is not found.
func (wallet *Wallet) AccountByPublicKey(pubKey string) (core.ValidatorAccount, error) {
	id, exists := wallet.indexMapper[pubKey]
	if !exists {
		return nil, ErrAccountNotFound
	}
	return wallet.AccountByID(id)
}
