package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var cfg Config
var replacer = strings.NewReplacer("-", "_")

var rootCmd = &cobra.Command{
	Use:   "digestabotctl",
	Short: "Update image hashes in your files",
}

type Config struct {
	FlagTypes []string `json:"flag_types" mapstructure:"flag_types"`
	Directory string   `json:"directory" mapstructure:"directory"`
	CreatePR  bool     `json:"create_pr" mapstructure:"create_pr"`
	Logger    *slog.Logger
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.digestabot.json)")
}

func initConfig() {

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("json")
		viper.SetConfigName(".digestabot")
	}

	viper.SetEnvPrefix("digestabot")
	viper.AutomaticEnv()
	// replace - with _ in env vars
	viper.SetEnvKeyReplacer(replacer)

	// If a config file is found, read it in.
	cfg.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := viper.ReadInConfig(); err == nil {
		cfg.Logger.Debug(fmt.Sprintf("using config %s", viper.ConfigFileUsed()))
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		cobra.CheckErr(err)
	}
}
