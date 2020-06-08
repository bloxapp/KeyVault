package wallet_hd

import (
	"fmt"
	"github.com/bloxapp/KeyVault/core"
	"github.com/google/uuid"
)

// according to https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2334.md
const (
	WithdrawalKeyPath = "/0"
	WithdrawalKeyName = "wallet_withdrawal_key_unique"
	ValidatorKeyPath = "/0/%d"
)

// an hierarchical deterministic wallet
type HDWallet struct {
	name 		string
	id 			uuid.UUID
	walletType 	core.WalletType
	key 		*core.DerivableKey // the node key from which all accounts are derived
	indexMapper map[string]uuid.UUID
	context 	*core.PortfolioContext
}

func NewHDWallet(name string, key *core.DerivableKey, context *core.PortfolioContext) *HDWallet {
	return &HDWallet{
		name:            name,
		id:              uuid.New(),
		walletType:      core.HDWallet,
		key:        	 key,
		indexMapper: 	 make(map[string]uuid.UUID),
		context:		 context,
	}
}

// ID provides the ID for the wallet.
func (wallet *HDWallet) ID() uuid.UUID {
	return wallet.id
}

// Name provides the name for the wallet.
func (wallet *HDWallet) Name() string {
	return wallet.name
}

// Type provides the type of the wallet.
func (wallet *HDWallet) Type() core.WalletType {
	return wallet.walletType
}

// GetWithdrawalAccount returns this wallet's withdrawal key pair in the wallet as described in EIP-2334.
// This will error if an account with the name already exists.
func (wallet *HDWallet) GetWithdrawalAccount() (core.Account, error) {
	account,err := wallet.AccountByName(WithdrawalKeyName)
	if err != nil {
		return nil,err
	}

	if account == nil { // create on the fly
		created,err := wallet.createKey(WithdrawalKeyName,WithdrawalKeyPath,core.WithdrawalAccount)
		if err != nil {
			return nil,err
		}
		return created,nil
	}

	return wallet.AccountByName(WithdrawalKeyName)
}

// CreateValidatorKey creates a new validation (validator) key pair in the wallet.
// This will error if an account with the name already exists.
func (wallet *HDWallet) CreateValidatorAccount(name string) (core.Account, error) {
	path := fmt.Sprintf(ValidatorKeyPath,len(wallet.indexMapper))
	return wallet.createKey(name,path,core.ValidatorAccount)
}

// Accounts provides all accounts in the wallet.
func (wallet *HDWallet) Accounts() <-chan core.Account {
	ch := make (chan core.Account,1024) // TODO - handle more? change from chan?
	go func() {
		for name := range wallet.indexMapper {
			id := wallet.indexMapper[name]
			account,err := wallet.AccountByID(id)
			if err != nil {
				continue
			}
			ch <- account
		}
		close(ch)
	}()

	return ch
}

// AccountByID provides a single account from the wallet given its ID.
// This will error if the account is not found.
func (wallet *HDWallet) AccountByID(id uuid.UUID) (core.Account, error) {
	return wallet.context.Storage.OpenAccount(id)
}

func (wallet *HDWallet) SetContext(ctx *core.PortfolioContext) {
	wallet.context = ctx
}

// AccountByName provides a single account from the wallet given its name.
// This will error if the account is not found.
func (wallet *HDWallet) AccountByName(name string) (core.Account, error) {
	id := wallet.indexMapper[name]
	return wallet.AccountByID(id)
}

func (wallet *HDWallet) createKey(name string, path string, accountType core.AccountType) (core.Account, error) {
	var retAccount *HDAccount

	// create account
	key,err := wallet.key.Derive(path)
	if err != nil {
		return nil,err
	}
	newContext := wallet.context.CopyForAccount(wallet.ID())
	retAccount,err = newHDAccount(
		name,
		accountType,
		key,
		newContext,
	)

	// register new wallet and save portfolio
	reset := func() {
		delete(wallet.indexMapper,name)
	}
	wallet.indexMapper[name] = retAccount.ID()
	err = wallet.context.Storage.SaveAccount(retAccount)
	if err != nil {
		reset()
		return nil,err
	}
	err = wallet.context.Storage.SaveWallet(wallet)
	if err != nil {
		reset()
		return nil,err
	}

	return retAccount,nil
}