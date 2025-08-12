package cmd

import (
	"github.com/spf13/cobra"
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

	return nil
}
