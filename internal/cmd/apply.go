package cmd

import (
	"context"
	"fmt"

	"go-unraid-clean/internal/apply"
	"go-unraid-clean/internal/config"
	"go-unraid-clean/internal/interactive"
	"go-unraid-clean/internal/report"

	"github.com/spf13/cobra"
)

var applyIn string
var applyConfirm bool
var applyInteractive bool

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a previously generated cleanup report",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}

		rep, err := report.ReadJSON(applyIn)
		if err != nil {
			return err
		}

		if applyInteractive {
			return interactive.Run(ctx, configPath, cfg, rep)
		}

		summary := report.Summarize(rep)
		fmt.Printf("Items: %d\n", summary.Total)
		if len(summary.ByType) > 0 {
			fmt.Printf("By type: %v\n", summary.ByType)
		}
		if len(summary.ByReason) > 0 {
			fmt.Printf("By reason: %v\n", summary.ByReason)
		}

		if !applyConfirm {
			fmt.Println("Review complete. Re-run with --confirm to apply deletions.")
			return nil
		}

		return apply.Run(ctx, cfg, rep)
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringVar(&applyIn, "in", "review.json", "Input review report to apply")
	applyCmd.Flags().BoolVar(&applyConfirm, "confirm", false, "Actually delete items from Sonarr/Radarr")
	applyCmd.Flags().BoolVarP(&applyInteractive, "interactive", "i", false, "Interactively review and apply actions per item")
}
