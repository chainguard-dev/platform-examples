package cmd

import (
	"fmt"
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

func init() {
	updateCmd.AddCommand(filesCmd)
}

func files(cmd *cobra.Command, args []string) error {

	fileTypes := viper.GetStringSlice("file_types")
	dir := viper.GetString("directory")

	for _, fileType := range fileTypes {
		matches, err := filepath.Glob(fmt.Sprintf("%s/%s", dir, fileType))
		if err != nil {
			return err
		}

		if len(matches) == 0 {
			continue
		}

		if err := digestabot.UpdateFiles(matches, cfg.Logger); err != nil {
			return err
		}
	}

	if viper.GetBool("create_pr") {
		opts := versioncontrol.CommitOptions{
			Directory: ".",
			Message:   "update digest hashes",
			Name:      "Test",
			Email:     "test@test.com",
			When:      time.Now(),
		}

		commit, err := versioncontrol.Commit(opts)
		if err != nil {
			return err
		}
		pr := platforms.PullRequest{
			Description: "this is a test",
			Diff:        commit,
		}

		platforms.NewGithubPR(pr)

	}

	return nil
}
