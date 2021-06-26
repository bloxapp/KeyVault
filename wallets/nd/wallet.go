package nd

import (
	"encoding/hex"
	"sort"
	"strconv"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/bloxapp/eth2-key-manager/core"
)

// Predefined errors
var (
	// ErrAccountNotFound is the error when account not found
	ErrAccountNotFound = errors.New("account not found")
)

// Wallet is hierarchical deterministic wallet
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
		walletType:  core.NDWallet,
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

// CreateValidatorAccount creates a new validation (validator) key pair in the wallet.
func (wallet *Wallet) CreateValidatorAccount(_ []byte, _ *int) (core.ValidatorAccount, error) {
	return nil, errors.Errorf("non deterministic wallet can't create validator, please use AddValidatorAccount")
}

// CreateValidatorAccountFromPrivateKey creates a new validation (validator) key pair in the wallet.
func (wallet *Wallet) CreateValidatorAccountFromPrivateKey(_ []byte, _ *int) (core.ValidatorAccount, error) {
	return nil, errors.Errorf("non deterministic wallet can't create validator, please use AddValidatorAccount")
}

// AddValidatorAccount adds the given account
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

// DeleteAccountByPublicKey deletes account by public key
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
