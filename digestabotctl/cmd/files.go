package cmd

import (
	"time"

	"github.com/chainguard-dev/platform-examples/digestabotctl/digestabot"
	"github.com/chainguard-dev/platform-examples/digestabotctl/platforms"
	"github.com/chainguard-dev/platform-examples/digestabotctl/versioncontrol"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var filesCmd = &cobra.Command{
	Use:          "files",
	Short:        "Update digest hashes in files",
	RunE:         files,
	PreRunE:      validateFiles,
	SilenceUsage: true,
}

var requiredPRFlags = []string{
	"owner",
	"repo",
	"branch",
	"token",
	"platform",
	"email",
}

func init() {
	updateCmd.AddCommand(filesCmd)
	fileFlags(filesCmd)
	bindFileFlags(filesCmd)
}

func validateFiles(cmd *cobra.Command, args []string) error {
	if !viper.GetBool("create_pr") {
		return nil
	}

	if err := validateEnvs(requiredPRFlags...); err != nil {
		return err
	}

	_, ok := platforms.ValidPlatforms[platforms.GitPlatform(viper.GetString("platform"))]
	if !ok {
		return platforms.ErrInvalidPlatform
	}

	return nil
}

func files(cmd *cobra.Command, args []string) error {
	opts := versioncontrol.CommitOptions{
		Directory: ".",
		Message:   viper.GetString("title"),
		When:      time.Now(),
		Branch:    viper.GetString("branch"),
		Token:     viper.GetString("token"),
		Signer:    signer,
		Name:      viper.GetString("name"),
		Email:     viper.GetString("email"),
	}

	checkout, err := versioncontrol.Checkout(opts)
	if err != nil {
		return err
	}

	fileTypes := viper.GetStringSlice("file_types")
	dir := viper.GetString("directory")

	files, err := digestabot.FindFiles(fileTypes, dir)
	if err != nil {
		return err
	}

	if err := digestabot.UpdateFiles(files, cfg.Logger); err != nil {
		return err
	}

	if viper.GetBool("create_pr") {
		platform := viper.GetString("platform")

		return handlePRForPlatform(platform, checkout, opts)
	}

	return nil
}

func handlePRForPlatform(platform string, checkout versioncontrol.CheckoutResponse, opts versioncontrol.CommitOptions) error {
	commit, err := versioncontrol.CommitAndPush(checkout.Repo, checkout.Worktree, opts)
	if err != nil {
		return err
	}

	pr := platforms.PullRequest{
		Description: viper.GetString("description"),
		Title:       viper.GetString("title"),
		Diff:        commit,
		Base:        viper.GetString("base"),
		Head:        viper.GetString("branch"),
		Labels:      viper.GetStringSlice("labels"),
		RepoData: platforms.RepoData{
			Repo:  viper.GetString("repo"),
			Owner: viper.GetString("owner"),
			Token: viper.GetString("token"),
		},
	}

	platformFunc := platforms.ValidPlatforms[platforms.GitPlatform(platform)]
	if platformFunc == nil {
		return platforms.ErrInvalidPlatform
	}

	creator, err := platformFunc(pr)
	if err != nil {
		return err
	}

	return creator.CreatePR(cfg.Logger)
}
