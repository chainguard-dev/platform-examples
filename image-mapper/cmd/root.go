package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "image-mapper",
	Short: "Map upstream image references to Chainguard images.",
}

func Execute() error {
	return rootCmd.Execute()
}
