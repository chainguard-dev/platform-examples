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

func init() {
	updateCmd.AddCommand(filesCmd)
}

func files(cmd *cobra.Command, args []string) error {

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
