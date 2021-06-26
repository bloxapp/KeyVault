package eth1deposit

import (
	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	ssz "github.com/prysmaticlabs/go-ssz"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/shared/params"
	util "github.com/wealdtech/go-eth2-util"

	"github.com/bloxapp/eth2-key-manager/core"
)

const (
	// MaxEffectiveBalanceInGwei is the max effective balance
	MaxEffectiveBalanceInGwei uint64 = 32000000000

	// BLSWithdrawalPrefixByte is the BLS withdrawal prefix
	BLSWithdrawalPrefixByte byte = byte(0)
)

// IsSupportedDepositNetwork returns true if the given network is supported
var IsSupportedDepositNetwork = func(network core.Network) bool {
	return network == core.PraterNetwork || network == core.MainNetwork || network == core.PyrmontNetwork
}

// DepositData is basically copied from https://github.com/prysmaticlabs/prysm/blob/master/shared/keystore/deposit_input.go
func DepositData(validationKey *core.HDKey, withdrawalPubKey []byte, network core.Network, amountInGwei uint64) (*ethpb.Deposit_Data, [32]byte, error) {
	if !IsSupportedDepositNetwork(network) {
		return nil, [32]byte{}, errors.Errorf("Network %s is not supported", network)
	}

	depositData := struct {
		PublicKey             []byte `ssz-size:"48"`
		WithdrawalCredentials []byte `ssz-size:"32"`
		Amount                uint64
	}{
		PublicKey:             validationKey.PublicKey().Serialize(),
		WithdrawalCredentials: withdrawalCredentialsHash(withdrawalPubKey),
		Amount:                amountInGwei,
	}
	objRoot, err := ssz.HashTreeRoot(depositData)
	if err != nil {
		return nil, [32]byte{}, errors.Wrap(err, "failed to determine the root hash of deposit data")
	}

	// Create domain
	domain, err := helpers.ComputeDomain(params.BeaconConfig().DomainDeposit, network.ForkVersion(), make([]byte, 32))
	if err != nil {
		return nil, [32]byte{}, errors.Wrap(err, "failed to calculate domain")
	}

	// Prepare for sig
	signingContainer := struct {
		Root   []byte `json:"object_root,omitempty" ssz-size:"32"`
		Domain []byte `json:"domain,omitempty" ssz-size:"32"`
	}{
		Root:   objRoot[:],
		Domain: domain,
	}
	root, err := ssz.HashTreeRoot(signingContainer)
	if err != nil {
		return nil, [32]byte{}, errors.Wrap(err, "failed to determine the root hash of signing container")
	}

	// Sign
	sig, err := validationKey.Sign(root[:])
	if err != nil {
		return nil, [32]byte{}, errors.Wrap(err, "failed to sign the root")
	}

	signedDepositData := &ethpb.Deposit_Data{
		PublicKey:             validationKey.PublicKey().Serialize(),
		WithdrawalCredentials: withdrawalCredentialsHash(withdrawalPubKey),
		Amount:                amountInGwei,
		Signature:             sig,
	}

	depositDataRoot, err := ssz.HashTreeRoot(signedDepositData)
	if err != nil {
		return nil, [32]byte{}, errors.Wrap(err, "failed to determine the root hash of deposit data")
	}

	return signedDepositData, depositDataRoot, nil
}

// withdrawalCredentialsHash forms a 32 byte hash of the withdrawal public
// address.
//
// The specification is as follows:
//   withdrawal_credentials[:1] == BLS_WITHDRAWAL_PREFIX_BYTE
//   withdrawal_credentials[1:] == hash(withdrawal_pubkey)[1:]
// where withdrawal_credentials is of type bytes32.
func withdrawalCredentialsHash(withdrawalPubKey []byte) []byte {
	h := util.SHA256(withdrawalPubKey)
	return append([]byte{BLSWithdrawalPrefixByte}, h[1:]...)[:32]
}
