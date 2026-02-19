package cmd

import (
	"context"

	"go-unraid-clean/internal/config"
	"go-unraid-clean/internal/interactive"
	"go-unraid-clean/internal/report"

	"github.com/spf13/cobra"
)

var interactiveIn string

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactively review and apply actions per item",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}

		rep, err := report.ReadJSON(interactiveIn)
		if err != nil {
			return err
		}

		return interactive.Run(ctx, configPath, cfg, rep)
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
	interactiveCmd.Flags().StringVar(&interactiveIn, "in", "review.json", "Input review report to apply")
}
