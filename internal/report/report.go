package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

type Report struct {
	GeneratedAt time.Time `json:"generated_at"`
	Items       []Item    `json:"items"`
}

type Item struct {
	Type           string     `json:"type"`
	Title          string     `json:"title"`
	RadarrID       *int       `json:"radarr_id,omitempty"`
	SonarrID       *int       `json:"sonarr_id,omitempty"`
	Path           string     `json:"path"`
	SizeBytes      int64      `json:"size_bytes"`
	AddedAt        *time.Time `json:"added_at,omitempty"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
	Reason         string     `json:"reason"`
}

type Summary struct {
	Total    int
	ByType   map[string]int
	ByReason map[string]int
}

func WriteJSON(path string, report *Report) error {
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}

	payload = append(payload, '\n')
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

func WriteCSV(path string, report *Report) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create csv: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{
		"type",
		"title",
		"radarr_id",
		"sonarr_id",
		"path",
		"size_bytes",
		"size_gib",
		"added_at",
		"last_activity_at",
		"reason",
	}); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for _, item := range report.Items {
		row := []string{
			item.Type,
			item.Title,
			formatOptionalInt(item.RadarrID),
			formatOptionalInt(item.SonarrID),
			item.Path,
			fmt.Sprintf("%d", item.SizeBytes),
			formatSizeGiB(item.SizeBytes),
			formatOptionalTime(item.AddedAt),
			formatOptionalTime(item.LastActivityAt),
			item.Reason,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}
	return nil
}

func ReadJSON(path string) (*Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read report: %w", err)
	}

	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parse report: %w", err)
	}
	return &report, nil
}

func Summarize(report *Report) Summary {
	out := Summary{
		Total:    len(report.Items),
		ByType:   map[string]int{},
		ByReason: map[string]int{},
	}
	for _, item := range report.Items {
		if item.Type != "" {
			out.ByType[item.Type]++
		}
		if item.Reason != "" {
			out.ByReason[item.Reason]++
		}
	}
	return out
}

func PrintTable(report *Report) {
	if len(report.Items) == 0 {
		fmt.Println("No items to review.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(w, "TYPE\tTITLE\tSIZE(GiB)\tADDED\tLAST_ACTIVITY\tREASON\tPATH")
	for _, item := range report.Items {
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Type,
			item.Title,
			formatSizeGiB(item.SizeBytes),
			formatOptionalTime(item.AddedAt),
			formatOptionalTime(item.LastActivityAt),
			item.Reason,
			item.Path,
		)
	}
	_ = w.Flush()
}

func formatOptionalInt(val *int) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%d", *val)
}

func formatOptionalTime(val *time.Time) string {
	if val == nil {
		return ""
	}
	return val.UTC().Format(time.RFC3339)
}

func formatSizeGiB(bytes int64) string {
	if bytes <= 0 {
		return "0"
	}
	gib := float64(bytes) / (1024 * 1024 * 1024)
	return fmt.Sprintf("%.2f", gib)
}
