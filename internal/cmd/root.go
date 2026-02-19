package cmd

import (
	"fmt"
	"os"

	"go-unraid-clean/internal/logging"

	"github.com/spf13/cobra"
)

var configPath string
var verbose bool

var rootCmd = &cobra.Command{
	Use:           "go-unraid-clean",
	Short:         "Review-and-apply cleanup for Plex, Sonarr, and Radarr",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logging.Setup(verbose)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "config.yaml", "Path to config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}
