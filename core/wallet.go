package core

import (
	"github.com/google/uuid"
)

// WalletType represents wallet type
type WalletType = string

// Wallet types
const (
	HDWallet WalletType = "HD" // hierarchical deterministic wallet
	NDWallet WalletType = "ND" // non - deterministic
)

// Wallet is a container of accounts.
// Accounts = key pairs
type Wallet interface {
	// ID provides the ID for the wallet.
	ID() uuid.UUID

	// Type provides the type of the wallet.
	Type() WalletType

	// CreateValidatorKey creates a new validation (validator) key pair in the wallet.
	// Keep in mind ND wallets will probably not allow this function, use AddValidatorAccount.
	CreateValidatorAccount(seed []byte, indexPointer *int) (ValidatorAccount, error)

	// Create validator account from Private Key
	CreateValidatorAccountFromPrivateKey(privateKey []byte, indexPointer *int) (ValidatorAccount, error)

	// Used to specifically add an account.
	// Keep in mind HD wallets will probably not allow this function, use CreateValidatorAccount.
	AddValidatorAccount(account ValidatorAccount) error

	// Accounts provides all accounts in the wallet.
	Accounts() []ValidatorAccount

	// AccountByID provides a single account from the wallet given its ID.
	// This will error if the account is not found.
	// should return account = nil if not found (not an error!)
	AccountByID(id uuid.UUID) (ValidatorAccount, error)

	// AccountByPublicKey provides a single account from the wallet given its public key.
	// This will error if the account is not found.
	// should return account = nil if not found (not an error!)
	AccountByPublicKey(pubKey string) (ValidatorAccount, error)

	// DeleteAccountByPublicKey delete an account from the wallet given its public key.
	// This will error if the account is not found.
	// should return nil if not error otherwise the error
	DeleteAccountByPublicKey(pubKey string) error

	// SetContext sets the given context
	SetContext(ctx *WalletContext)
}

// WalletContext represents the wallet's context type
type WalletContext struct {
	Storage Storage
}
