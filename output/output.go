package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
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

func FormatTaskTable(tasks []model.Task, projectByID map[int64]string, now time.Time) string {
	if now.IsZero() {
		now = time.Now()
	}

	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, "ID\tPriority\tStar\tTitle\tDue\tProject\tLabels")
	for _, task := range tasks {
		project := projectByID[task.ProjectID]
		star := ""
		if task.IsFavorite {
			star = "*"
		}

		fmt.Fprintf(
			tw,
			"%d\t%d\t%s\t%s\t%s\t%s\t%s\n",
			task.ID,
			task.Priority,
			star,
			task.Title,
			FormatDateText(task.DueDate, now),
			project,
			joinLabelTitles(task.Labels),
		)
	}

	_ = tw.Flush()
	return strings.TrimRight(b.String(), "\n")
}

func FormatKeyValues(items []KeyValue) string {
	if len(items) == 0 {
		return ""
	}

	maxKey := 0
	for _, item := range items {
		if len(item.Key) > maxKey {
			maxKey = len(item.Key)
		}
	}

	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%-*s : %s", maxKey, item.Key, item.Value))
	}

	return strings.Join(lines, "\n")
}

func FormatProjectTable(projects []model.Project) string {
	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, "ID\tTitle\tParent\tColor\tFavorite")
	for _, project := range projects {
		parent := ""
		if project.ParentProjectID != nil {
			parent = fmt.Sprintf("%d", *project.ParentProjectID)
		}

		star := ""
		if project.IsFavorite {
			star = "*"
		}

		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", project.ID, project.Title, parent, project.HexColor, star)
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

func joinLabelTitles(labels []model.Label) string {
	if len(labels) == 0 {
		return ""
	}

	titles := make([]string, 0, len(labels))
	for _, label := range labels {
		titles = append(titles, label.Title)
	}

	sort.Strings(titles)
	return strings.Join(titles, ",")
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
