package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/chainguard-dev/platform-examples/digestabotctl/digestabot"
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

		digestabot.UpdateFiles(matches, cfg.Logger)
	}

	return nil
}
