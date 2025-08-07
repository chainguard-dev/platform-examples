package cmd

import (
	"io/fs"
	"path/filepath"
	"time"

	"github.com/chainguard-dev/platform-examples/digestabotctl/digestabot"
	"github.com/chainguard-dev/platform-examples/digestabotctl/platforms"
	"github.com/chainguard-dev/platform-examples/digestabotctl/versioncontrol"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "Update digest hashes in files",
	RunE:  files,
}

var requiredPRFlags = []string{
	"owner",
	"repo",
	"branch",
	"token",
}

func init() {
	updateCmd.AddCommand(filesCmd)
}

func files(cmd *cobra.Command, args []string) error {
	opts := versioncontrol.CommitOptions{
		Directory: ".",
		Message:   viper.GetString("title"),
		When:      time.Now(),
		Branch:    viper.GetString("branch"),
		Token:     viper.GetString("token"),
	}

	checkout, err := versioncontrol.Checkout(opts)
	if err != nil {
		return err
	}

	fileTypes := viper.GetStringSlice("file_types")
	dir := viper.GetString("directory")
	files := []string{}

	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		for _, pattern := range fileTypes {
			matched, err := filepath.Match(pattern, base)
			if err != nil {
				return err
			}
			if matched {
				files = append(files, path)
				break
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if err := digestabot.UpdateFiles(files, cfg.Logger); err != nil {
		return err
	}

	if viper.GetBool("create_pr") {
		if err := validateEnvs(requiredPRFlags...); err != nil {
			return err
		}

		commit, err := versioncontrol.CommitAndPush(checkout.Repo, checkout.Worktree, opts)
		if err != nil {
			return err
		}
		gh := platforms.GitHub{
			Repo:  viper.GetString("repo"),
			Owner: viper.GetString("owner"),
			Token: viper.GetString("token"),
		}
		pr := platforms.PullRequest{
			Description: "this is a test",
			Title:       viper.GetString("title"),
			Diff:        commit,
			Base:        viper.GetString("base"),
			Head:        viper.GetString("branch"),
		}

		ghPR, err := platforms.NewGithubPR(gh, pr)
		if err != nil {
			return err
		}

		if err := gh.CreatePR(ghPR); err != nil {
			return err
		}

	}

	return nil
}
