package account

import (
	"github.com/spf13/cobra"

	rootcmd "github.com/bloxapp/eth2-key-manager/cli/cmd"
	"github.com/bloxapp/eth2-key-manager/cli/cmd/wallet/cmd/account/flag"
	"github.com/bloxapp/eth2-key-manager/cli/cmd/wallet/cmd/account/handler"
)

// createCmd represents the create account command.
var createSeedlessCmd = &cobra.Command{
	Use:   "create-seedless",
	Short: "Creates a wallet account.",
	Long:  `This command creates an account using private key.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		handler := handler.New(rootcmd.ResultPrinter)
		return handler.CreateSeedless(cmd, args)
	},
}

func init() {
	// Define flags for the command.
	flag.AddIndexFlag(createSeedlessCmd)
	flag.AddPrivateKeyFlag(createSeedlessCmd)
	flag.AddAccumulateFlag(createSeedlessCmd)
	flag.AddResponseTypeFlag(createSeedlessCmd)
	flag.AddHighestSourceFlag(createSeedlessCmd)
	flag.AddHighestTargetFlag(createSeedlessCmd)
	flag.AddHighestProposalFlag(createSeedlessCmd)
	rootcmd.AddNetworkFlag(createSeedlessCmd)

	Command.AddCommand(createSeedlessCmd)
}
