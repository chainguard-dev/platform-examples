package cmd

import (
	"github.com/chainguard-dev/platform-examples/digestabotctl/versioncontrol"
	"github.com/go-git/go-git/v6"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:               "update",
	Short:             "Command to control updates to digests",
	PersistentPreRunE: updatePreRunE,
}

var requiredUpdateFlags = []string{
	"branch",
}

var signer git.Signer

func init() {
	rootCmd.AddCommand(updateCmd)
	prFlags(updateCmd)

}

func updatePreRunE(cmd *cobra.Command, args []string) error {
	//bind flags for Viper
	bindPRFlags(cmd)

	if err := validateEnvs(requiredUpdateFlags...); err != nil {
		return err
	}
	if viper.GetBool("sign") {
		var err error
		signer, err = versioncontrol.NewSigner(cmd.Context())
		if err != nil {
			return err
		}
	}
	return nil
}
