package cmd

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/chainguard-dev/platform-examples/digestabotctl/digestabot"
	"github.com/chainguard-dev/platform-examples/digestabotctl/platforms"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func validateEnvs(vals ...string) error {
	var errs []error
	for _, v := range vals {
		if !viper.IsSet(v) {
			errs = append(errs, fmt.Errorf("%v must be set", v))
		}
	}

	return errors.Join(errs...)
}

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
	cmd.Flags().StringSliceP("file-types", "f", digestabot.DefaultFileTypes, "Files to update")
	cmd.Flags().StringP("directory", "d", ".", "Directory to update files")
}

// bindFileFlags binds the pr flag values to viper
func bindPRFlags(cmd *cobra.Command) {
	viper.BindPFlag("create_pr", cmd.Flags().Lookup("create-pr"))
	viper.BindPFlag("owner", cmd.Flags().Lookup("owner"))
	viper.BindPFlag("repo", cmd.Flags().Lookup("repo"))
	viper.BindPFlag("branch", cmd.Flags().Lookup("branch"))
	viper.BindPFlag("base", cmd.Flags().Lookup("base"))
	viper.BindPFlag("title", cmd.Flags().Lookup("title"))
	viper.BindPFlag("token", cmd.Flags().Lookup("token"))
	viper.BindPFlag("description", cmd.Flags().Lookup("description"))
	viper.BindPFlag("platform", cmd.Flags().Lookup("platform"))
	viper.BindPFlag("sign", cmd.Flags().Lookup("sign"))
	viper.BindPFlag("signing-token", cmd.Flags().Lookup("signing-token"))
	viper.BindPFlag("name", cmd.Flags().Lookup("name"))
	viper.BindPFlag("email", cmd.Flags().Lookup("email"))
	viper.BindPFlag("labels", cmd.Flags().Lookup("labels"))
}

// prFlags adds the pr flags to the passed in command
func prFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("create-pr", false, "Create a PR")
	cmd.PersistentFlags().String("owner", "", "Repo owner/organization")
	cmd.PersistentFlags().String("repo", "", "Repo name")
	cmd.PersistentFlags().String("branch", "", "branch for commit")
	cmd.PersistentFlags().String("base", "main", "branch for PR to merge into")
	cmd.PersistentFlags().String("title", "Updating image digests", "PR title")
	cmd.PersistentFlags().String("token", "", "API token")
	cmd.PersistentFlags().String("description", "Updating image digests", "PR description")
	cmd.PersistentFlags().String("platform", "", fmt.Sprintf("Platform to create the PR. Options are %s", slices.Collect(maps.Keys(platforms.ValidPlatforms))))
	cmd.PersistentFlags().Bool("sign", false, "Sign the commit")
	cmd.PersistentFlags().String("signing-token", "", "OIDC token for signing commit")
	cmd.PersistentFlags().String("name", "digestabotctl", "Name for commit")
	cmd.PersistentFlags().String("email", "", "Email for commit")
	cmd.PersistentFlags().StringSlice("labels", []string{"automated pr", "kind/cleanup", "release-note-none"}, "Labels to apply to the PR")
}
