package cmd

import (
	"context"
	"fmt"

	"go-unraid-clean/internal/config"
	"go-unraid-clean/internal/report"
	"go-unraid-clean/internal/scan"

	"github.com/spf13/cobra"
)

var scanOut string
var scanCSV string
var scanTable bool
var scanSort string
var scanOrder string

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Generate a cleanup review report",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}

		rep, err := scan.Run(ctx, cfg, scan.Options{
			SortBy:    scanSort,
			SortOrder: scanOrder,
		})
		if err != nil {
			return err
		}

		if scanOut != "" {
			if err := report.WriteJSON(scanOut, rep); err != nil {
				return err
			}
			fmt.Printf("Wrote report to %s (%d items)\n", scanOut, len(rep.Items))
		}

		if scanCSV != "" {
			if err := report.WriteCSV(scanCSV, rep); err != nil {
				return err
			}
			fmt.Printf("Wrote CSV to %s (%d items)\n", scanCSV, len(rep.Items))
		}

		if scanTable {
			report.PrintTable(rep)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().StringVar(&scanOut, "out", "review.json", "Output path for the review report")
	scanCmd.Flags().StringVar(&scanCSV, "csv", "", "Optional CSV output path for review")
	scanCmd.Flags().BoolVar(&scanTable, "table", false, "Print a pretty table of results to stdout")
	scanCmd.Flags().StringVar(&scanSort, "sort", "size", "Sort by: size, added, gap, last_activity, inactivity")
	scanCmd.Flags().StringVar(&scanOrder, "order", "desc", "Sort order: asc or desc")
}
