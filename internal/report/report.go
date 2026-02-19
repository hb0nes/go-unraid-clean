package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

type Report struct {
	GeneratedAt time.Time `json:"generated_at"`
	Items       []Item    `json:"items"`
}

type Item struct {
	Type               string      `json:"type"`
	Title              string      `json:"title"`
	RadarrID           *int        `json:"radarr_id,omitempty"`
	SonarrID           *int        `json:"sonarr_id,omitempty"`
	TMDBID             *int        `json:"tmdb_id,omitempty"`
	TVDBID             *int        `json:"tvdb_id,omitempty"`
	IMDBID             string      `json:"imdb_id,omitempty"`
	Path               string      `json:"path"`
	SizeBytes          int64       `json:"size_bytes"`
	AddedAt            *time.Time  `json:"added_at,omitempty"`
	FirstActivityAt    *time.Time  `json:"first_activity_at,omitempty"`
	LastActivityAt     *time.Time  `json:"last_activity_at,omitempty"`
	TopUsers           []UserWatch `json:"top_users,omitempty"`
	TopUsersTotalHours float64     `json:"top_users_total_hours,omitempty"`
	TotalWatchHours    float64     `json:"total_watch_hours,omitempty"`
	SeriesStatus       string      `json:"series_status,omitempty"`
	Reason             string      `json:"reason"`
}

type UserWatch struct {
	User  string  `json:"user"`
	Hours float64 `json:"hours"`
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
		"series_status",
		"path",
		"size_bytes",
		"size_gib",
		"added_at",
		"first_activity_at",
		"last_activity_at",
		"gap_days",
		"inactivity_days",
		"top_users",
		"top_users_hours_total",
		"total_watch_hours",
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
			item.SeriesStatus,
			item.Path,
			fmt.Sprintf("%d", item.SizeBytes),
			formatSizeGiB(item.SizeBytes),
			formatOptionalTime(item.AddedAt),
			formatOptionalTime(item.FirstActivityAt),
			formatOptionalTime(item.LastActivityAt),
			formatGapDays(item.AddedAt, item.FirstActivityAt, report.GeneratedAt),
			formatInactivityDays(item.AddedAt, item.LastActivityAt, report.GeneratedAt),
			formatTopUsers(item.TopUsers, item.TopUsersTotalHours),
			formatHours(item.TopUsersTotalHours),
			formatHours(item.TotalWatchHours),
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
	fmt.Fprintln(w, "TYPE\tTITLE\tSTATUS\tSIZE(GiB)\tADDED\tFIRST_ACTIVITY\tLAST_ACTIVITY\tGAP_DAYS\tINACTIVITY_DAYS\tWATCH_HOURS\tTOP_USERS\tREASON\tPATH")
	for _, item := range report.Items {
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Type,
			item.Title,
			item.SeriesStatus,
			formatSizeGiB(item.SizeBytes),
			formatOptionalTime(item.AddedAt),
			formatOptionalTime(item.FirstActivityAt),
			formatOptionalTime(item.LastActivityAt),
			formatGapDays(item.AddedAt, item.FirstActivityAt, report.GeneratedAt),
			formatInactivityDays(item.AddedAt, item.LastActivityAt, report.GeneratedAt),
			formatHours(item.TotalWatchHours),
			formatTopUsers(item.TopUsers, item.TopUsersTotalHours),
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

func formatGapDays(addedAt *time.Time, firstActivityAt *time.Time, generatedAt time.Time) string {
	if addedAt == nil {
		return ""
	}
	end := generatedAt
	if firstActivityAt != nil {
		end = *firstActivityAt
	}
	span := end.Sub(*addedAt).Hours() / 24
	if span < 0 {
		span = 0
	}
	return fmt.Sprintf("%.1f", span)
}

func formatInactivityDays(addedAt *time.Time, lastActivityAt *time.Time, generatedAt time.Time) string {
	end := generatedAt
	if lastActivityAt != nil {
		end = *lastActivityAt
	} else if addedAt == nil {
		return ""
	}
	span := generatedAt.Sub(end).Hours() / 24
	if span < 0 {
		span = 0
	}
	return fmt.Sprintf("%.1f", span)
}

func formatTopUsers(users []UserWatch, total float64) string {
	if len(users) == 0 {
		return ""
	}
	parts := make([]string, 0, len(users)+1)
	for _, user := range users {
		parts = append(parts, fmt.Sprintf("%s:%.1fh", user.User, user.Hours))
	}
	if total > 0 {
		parts = append(parts, fmt.Sprintf("total:%.1fh", total))
	}
	return strings.Join(parts, " ")
}

func formatHours(hours float64) string {
	if hours <= 0 {
		return ""
	}
	return fmt.Sprintf("%.1f", hours)
}
