package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/p3psi-boo/vikunja-cli/model"
)

type KeyValue struct {
	Key   string
	Value string
}

func WriteJSONSingle(w io.Writer, value any) error {
	return writeJSON(w, value)
}

func WriteJSONList[T any](w io.Writer, values []T) error {
	return writeJSON(w, values)
}

// TaskTableOptions tunes how FormatTaskTable renders rows.
type TaskTableOptions struct {
	AbsoluteDate bool // show "2026-06-29 (in 2d)" instead of just "in 2d"
}

func FormatTaskTable(tasks []model.Task, projectByID map[int64]string, now time.Time) string {
	return FormatTaskTableWithOptions(tasks, projectByID, now, TaskTableOptions{})
}

func FormatTaskTableWithOptions(tasks []model.Task, projectByID map[int64]string, now time.Time, opts TaskTableOptions) string {
	if now.IsZero() {
		now = time.Now()
	}
	if projectByID == nil {
		projectByID = map[int64]string{}
	}

	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, "Done\tID\tPriority\tFav\tTitle\tDue\tProject\tLabels")
	for _, task := range tasks {
		project := projectByID[task.ProjectID]

		doneMark := ""
		if task.Done {
			doneMark = CheckMark()
		}

		favMark := ""
		if task.IsFavorite {
			favMark = FavoriteMark()
		}

		title := task.Title
		if task.Done {
			title = DoneStyle(title)
		}

		fmt.Fprintf(
			tw,
			"%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			doneMark,
			task.ID,
			formatPriorityCell(task.Priority),
			favMark,
			title,
			FormatDueCell(task.DueDate, task.Done, now, opts.AbsoluteDate),
			project,
			formatLabelCell(task.Labels),
		)
	}

	_ = tw.Flush()
	return strings.TrimRight(b.String(), "\n")
}

func formatPriorityCell(priority int) string {
	if priority == 0 {
		return "0"
	}
	return PriorityStyle(priority, strconv.Itoa(priority))
}

func formatLabelCell(labels []model.Label) string {
	if len(labels) == 0 {
		return ""
	}
	titles := make([]string, 0, len(labels))
	for _, label := range labels {
		titles = append(titles, label.Title)
	}
	sort.Strings(titles)
	joined := strings.Join(titles, ",")
	return Label(joined)
}

// FormatDueCell renders a due-date table cell, applying color based on how
// urgent the date is relative to now. Completed tasks get muted output.
func FormatDueCell(value model.NullableTime, done bool, now time.Time, absolute bool) string {
	if !value.Valid {
		return ""
	}

	if now.IsZero() {
		now = time.Now()
	}

	delta := value.Time.Sub(now)
	relative := formatRelative(value.Time, now)
	if done {
		return DoneStyle(relative)
	}

	// color thresholds
	switch {
	case delta < 0:
		relative = Overdue(relative)
	case delta < 24*time.Hour:
		relative = DueSoon(relative)
	}

	if !absolute {
		return relative
	}
	return value.Time.Format("2006-01-02") + " " + Muted("("+relative+")")
}

func FormatKeyValues(items []KeyValue) string {
	return formatKeyValues(items, false)
}

// FormatKeyValuesOmitEmpty behaves like FormatKeyValues but drops entries
// whose value is the empty string. Used by detail views to avoid printing a
// wall of empty Description / Favorite fields.
func FormatKeyValuesOmitEmpty(items []KeyValue) string {
	return formatKeyValues(items, true)
}

func formatKeyValues(items []KeyValue, omitEmpty bool) string {
	if len(items) == 0 {
		return ""
	}

	filtered := make([]KeyValue, 0, len(items))
	for _, item := range items {
		if omitEmpty && strings.TrimSpace(item.Value) == "" {
			continue
		}
		filtered = append(filtered, item)
	}
	if len(filtered) == 0 {
		return ""
	}

	maxKey := 0
	for _, item := range filtered {
		if len(item.Key) > maxKey {
			maxKey = len(item.Key)
		}
	}

	lines := make([]string, 0, len(filtered))
	for _, item := range filtered {
		lines = append(lines, fmt.Sprintf("%-*s : %s", maxKey, item.Key, item.Value))
	}

	return strings.Join(lines, "\n")
}

// EmptyMessage returns the human-friendly line printed when a list query
// yields no rows. filtered indicates whether the user applied any filters.
func EmptyMessage(resource string, filtered bool) string {
	if filtered {
		return "No " + resource + " matching the current filters."
	}
	return "No " + resource + " found."
}

func FormatProjectTable(projects []model.Project) string {
	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, "ID\tTitle\tParent\tColor\tFav")
	for _, project := range projects {
		parent := ""
		if project.ParentProjectID != nil {
			parent = fmt.Sprintf("%d", *project.ParentProjectID)
		}

		fav := ""
		if project.IsFavorite {
			fav = FavoriteMark()
		}

		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", project.ID, project.Title, parent, project.HexColor, fav)
	}

	_ = tw.Flush()
	return strings.TrimRight(b.String(), "\n")
}

func FormatLabelTable(labels []model.Label) string {
	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, "ID\tTitle\tColor")
	for _, label := range labels {
		fmt.Fprintf(tw, "%d\t%s\t%s\n", label.ID, label.Title, label.HexColor)
	}

	_ = tw.Flush()
	return strings.TrimRight(b.String(), "\n")
}

func FormatDateText(value model.NullableTime, now time.Time) string {
	if !value.Valid {
		return ""
	}

	if now.IsZero() {
		now = time.Now()
	}

	return formatRelative(value.Time, now)
}

// FormatDateRich renders an absolute date plus a parenthesized relative hint,
// e.g. "2026-06-29 (in 2d)". Overdue dates are colored; the caller decides
// whether to use this for detail views.
func FormatDateRich(value model.NullableTime, now time.Time) string {
	if !value.Valid {
		return ""
	}
	if now.IsZero() {
		now = time.Now()
	}
	abs := value.Time.Format("2006-01-02")
	rel := formatRelative(value.Time, now)
	if value.Time.Before(now) {
		rel = Overdue(rel)
	}
	return abs + " " + Muted("("+rel+")")
}

// DueState classifies a due date relative to now for summary counters.
type DueState int

const (
	DueStateNone DueState = iota
	DueStateFuture
	DueStateSoon
	DueStateOverdue
)

// ClassifyDue returns the urgency bucket of a due date.
func ClassifyDue(value model.NullableTime, now time.Time) DueState {
	if !value.Valid {
		return DueStateNone
	}
	if now.IsZero() {
		now = time.Now()
	}
	delta := value.Time.Sub(now)
	switch {
	case delta < 0:
		return DueStateOverdue
	case delta < 24*time.Hour:
		return DueStateSoon
	default:
		return DueStateFuture
	}
}

func FormatDateJSON(value model.NullableTime) string {
	if !value.Valid {
		return ""
	}

	return value.Time.UTC().Format(time.RFC3339)
}

func PrintInfo(w io.Writer, quiet bool, format string, args ...any) error {
	if quiet {
		return nil
	}

	message := fmt.Sprintf(format, args...)
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}

	_, err := io.WriteString(w, message)
	return err
}

func writeJSON(w io.Writer, value any) error {
	encoded, err := json.Marshal(normalizeJSONValue(value))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w, string(encoded))
	return err
}

func normalizeJSONValue(value any) any {
	switch v := value.(type) {
	case model.Project:
		return toJSONProject(v)
	case *model.Project:
		if v == nil {
			return nil
		}
		return toJSONProject(*v)
	case []model.Project:
		out := make([]map[string]any, 0, len(v))
		for _, item := range v {
			out = append(out, toJSONProject(item))
		}
		return out
	case model.Label:
		return toJSONLabel(v)
	case *model.Label:
		if v == nil {
			return nil
		}
		return toJSONLabel(*v)
	case []model.Label:
		out := make([]map[string]any, 0, len(v))
		for _, item := range v {
			out = append(out, toJSONLabel(item))
		}
		return out
	case model.Task:
		return toJSONTask(v)
	case *model.Task:
		if v == nil {
			return nil
		}
		return toJSONTask(*v)
	case []model.Task:
		out := make([]map[string]any, 0, len(v))
		for _, item := range v {
			out = append(out, toJSONTask(item))
		}
		return out
	case model.User:
		return map[string]any{
			"id":       v.ID,
			"username": v.Username,
			"name":     v.Name,
			"email":    v.Email,
			"created":  optionalRFC3339(v.Created),
			"updated":  optionalRFC3339(v.Updated),
		}
	default:
		return value
	}
}

func toJSONProject(project model.Project) map[string]any {
	return map[string]any{
		"id":                project.ID,
		"title":             project.Title,
		"description":       project.Description,
		"parent_project_id": project.ParentProjectID,
		"hex_color":         project.HexColor,
		"is_favorite":       project.IsFavorite,
		"created":           optionalRFC3339(project.Created),
		"updated":           optionalRFC3339(project.Updated),
	}
}

func toJSONLabel(label model.Label) map[string]any {
	return map[string]any{
		"id":          label.ID,
		"title":       label.Title,
		"description": label.Description,
		"hex_color":   label.HexColor,
		"created":     optionalRFC3339(label.Created),
		"updated":     optionalRFC3339(label.Updated),
	}
}

func toJSONTask(task model.Task) map[string]any {
	reminders := make([]map[string]any, 0, len(task.Reminders))
	for _, reminder := range task.Reminders {
		reminders = append(reminders, map[string]any{
			"id":              reminder.ID,
			"reminder":        optionalRFC3339(reminder.Reminder),
			"relative_period": reminder.RelativePeriod,
			"relative_to":     reminder.RelativeTo,
		})
	}

	labels := make([]map[string]any, 0, len(task.Labels))
	for _, label := range task.Labels {
		labels = append(labels, toJSONLabel(label))
	}

	return map[string]any{
		"id":          task.ID,
		"title":       task.Title,
		"description": task.Description,
		"done":        task.Done,
		"done_at":     optionalRFC3339(task.DoneAt),
		"due_date":    optionalRFC3339(task.DueDate),
		"reminders":   reminders,
		"project_id":  task.ProjectID,
		"priority":    task.Priority,
		"labels":      labels,
		"is_favorite": task.IsFavorite,
		"created":     optionalRFC3339(task.Created),
		"updated":     optionalRFC3339(task.Updated),
	}
}

func optionalRFC3339(value model.NullableTime) any {
	if !value.Valid {
		return nil
	}

	return value.Time.UTC().Format(time.RFC3339)
}

func formatRelative(date time.Time, now time.Time) string {
	delta := date.Sub(now)
	abs := delta
	if abs < 0 {
		abs = -abs
	}

	if abs < time.Minute {
		return "now"
	}

	units := []struct {
		name string
		dur  time.Duration
	}{
		{name: "y", dur: 365 * 24 * time.Hour},
		{name: "mo", dur: 30 * 24 * time.Hour},
		{name: "w", dur: 7 * 24 * time.Hour},
		{name: "d", dur: 24 * time.Hour},
		{name: "h", dur: time.Hour},
		{name: "m", dur: time.Minute},
	}

	for _, unit := range units {
		if abs >= unit.dur {
			count := int(abs / unit.dur)
			if delta > 0 {
				return fmt.Sprintf("in %d%s", count, unit.name)
			}
			return fmt.Sprintf("%d%s ago", count, unit.name)
		}
	}

	return "now"
}
