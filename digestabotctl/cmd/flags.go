package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Flags are defined here. Because of the way Viper binds values, if the same flag name is called
// with viper.BindPFlag multiple times during init() the value will be overwritten. For example if
// two subcommands each have a flag called name but they each have their own default values,
// viper can overwrite any value passed in for one subcommand with the default value of the other subcommand.
// The answer here is to not use init() and instead use something like PersistentPreRun to bind the
// viper values. Using init for the cobra flags is ok, they are only in here to limit duplication of names.

// bindFileFlags binds the file flag values to viper
func bindFileFlags(cmd *cobra.Command) {
	viper.BindPFlag("file_types", cmd.Flags().Lookup("file-types"))
	viper.BindPFlag("directory", cmd.Flags().Lookup("directory"))
}

// fileFlags adds the file flags to the passed in command
func fileFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceP("file-types", "f", []string{`*.yaml`, `*.yml`, `*.sh`, `*.tf`, `*.tfvars`, `Dockerfile*`, `Makefile*`}, "Files to update")
	cmd.PersistentFlags().StringP("directory", "d", ".", "Directory to update files")
}

// bindFileFlags binds the pr flag values to viper
func bindPRFlags(cmd *cobra.Command) {
	viper.BindPFlag("create_pr", cmd.Flags().Lookup("create-pr"))
}

// prFlags adds the pr flags to the passed in command
func prFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("create-pr", false, "Create a PR")
}
