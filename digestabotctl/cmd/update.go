package cmd

import (
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:              "update",
	Short:            "Command to control updates to digests",
	PersistentPreRun: bindUpdateCmdFlags,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	prFlags(updateCmd)

}
func bindUpdateCmdFlags(cmd *cobra.Command, args []string) {
	bindPRFlags(cmd)
}
