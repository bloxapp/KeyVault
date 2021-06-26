package handler

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/bloxapp/eth2-key-manager/core"
)

// CreateSeedless creates a new wallet account and prints the storage.
func (h *Account) CreateSeedless(cmd *cobra.Command, args []string) error {
	err := core.InitBLS()
	if err != nil {
		return errors.Wrap(err, "failed to init BLS")
	}

	accountFlags, err := CollectAccountFlags(cmd, true)
	if err != nil {
		return err
	}

	err = h.BuildAndPrintAccounts(accountFlags, true)
	if err != nil {
		return err
	}
	return nil
}
