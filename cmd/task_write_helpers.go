package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/dateparse"
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

func parseTaskIDs(args []string) ([]int64, error) {
	ids := make([]int64, 0, len(args))
	for _, arg := range args {
		id, err := strconv.ParseInt(strings.TrimSpace(arg), 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("invalid task id %q", arg)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func parseProjectID(ctx context.Context, client *api.Client, value string) (int64, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return 0, fmt.Errorf("--project is required")
	}

	return resolveProjectID(ctx, client, raw)
}

func resolveProjectID(ctx context.Context, client *api.Client, raw string) (int64, error) {
	if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
		return id, nil
	}

	projects, err := client.GetProjects(ctx)
	if err != nil {
		return 0, err
	}

	needle := strings.ToLower(strings.TrimSpace(raw))

	// 1. case-insensitive exact match
	if id, ok := uniqueMatch(projects, func(p model.Project) string { return p.Title }, needle, exact); ok {
		return id, nil
	}

	// 2. prefix match
	if id, ok := uniqueMatch(projects, func(p model.Project) string { return p.Title }, needle, prefix); ok {
		return id, nil
	}

	// 3. substring match
	candidates := collectMatches(projects, func(p model.Project) string { return p.Title }, needle, substring)
	if len(candidates) > 0 {
		return 0, ambiguousProjectError(raw, candidates)
	}

	return 0, fmt.Errorf("project %q not found. Run `vja project ls` to see available projects", raw)
}

type projectCandidate struct {
	ID    int64
	Title string
}

func ambiguousProjectError(raw string, candidates []projectCandidate) error {
	listed := make([]string, 0, len(candidates))
	for _, c := range candidates {
		listed = append(listed, fmt.Sprintf("%s (%d)", c.Title, c.ID))
	}
	return fmt.Errorf("multiple projects match %q: %s. Use the ID", raw, strings.Join(listed, ", "))
}

func parseLabelIDs(ctx context.Context, client *api.Client, values []string) ([]int64, error) {
	if len(values) == 0 {
		return nil, nil
	}

	ids := make([]int64, 0, len(values))
	var labels []model.Label
	labelsLoaded := false

	for _, value := range values {
		raw := strings.TrimSpace(value)
		if raw == "" {
			continue
		}

		if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
			ids = append(ids, id)
			continue
		}

		if !labelsLoaded {
			var err error
			labels, err = client.GetLabels(ctx)
			if err != nil {
				return nil, err
			}
			labelsLoaded = true
		}

		id, err := resolveLabelID(raw, labels)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func resolveLabelID(raw string, labels []model.Label) (int64, error) {
	needle := strings.ToLower(strings.TrimSpace(raw))

	if id, ok := uniqueLabelMatch(labels, needle, exact); ok {
		return id, nil
	}
	if id, ok := uniqueLabelMatch(labels, needle, prefix); ok {
		return id, nil
	}

	candidates := collectLabelMatches(labels, needle, substring)
	if len(candidates) > 0 {
		listed := make([]string, 0, len(candidates))
		for _, c := range candidates {
			listed = append(listed, fmt.Sprintf("%q (%d)", c.Title, c.ID))
		}
		return 0, fmt.Errorf("multiple labels match %q: %s. Use the ID", raw, strings.Join(listed, ", "))
	}

	return 0, fmt.Errorf("label %q not found. Run `vja label ls` to see available labels", raw)
}

type matchKind int

const (
	exact matchKind = iota
	prefix
	substring
)

func matchScore(haystack, needle string, kind matchKind) bool {
	switch kind {
	case exact:
		return haystack == needle
	case prefix:
		return strings.HasPrefix(haystack, needle)
	case substring:
		return strings.Contains(haystack, needle)
	}
	return false
}

func normalized(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// uniqueMatch returns the single project id that matches the needle under the
// given matchKind (case-insensitive). It returns ok=false when there are zero
// or multiple matches.
func uniqueMatch(projects []model.Project, titleOf func(model.Project) string, needle string, kind matchKind) (int64, bool) {
	var matchID int64
	matches := 0
	for _, p := range projects {
		if p.ID <= 0 {
			continue
		}
		if matchScore(normalized(titleOf(p)), needle, kind) {
			matchID = p.ID
			matches++
		}
	}
	if matches == 1 {
		return matchID, true
	}
	return 0, false
}

func collectMatches(projects []model.Project, titleOf func(model.Project) string, needle string, kind matchKind) []projectCandidate {
	var out []projectCandidate
	for _, p := range projects {
		if p.ID <= 0 {
			continue
		}
		if matchScore(normalized(titleOf(p)), needle, kind) {
			out = append(out, projectCandidate{ID: p.ID, Title: p.Title})
		}
	}
	return out
}

func uniqueLabelMatch(labels []model.Label, needle string, kind matchKind) (int64, bool) {
	var matchID int64
	matches := 0
	for _, l := range labels {
		if l.ID <= 0 {
			continue
		}
		if matchScore(normalized(l.Title), needle, kind) {
			matchID = l.ID
			matches++
		}
	}
	if matches == 1 {
		return matchID, true
	}
	return 0, false
}

func collectLabelMatches(labels []model.Label, needle string, kind matchKind) []projectCandidate {
	var out []projectCandidate
	for _, l := range labels {
		if l.ID <= 0 {
			continue
		}
		if matchScore(normalized(l.Title), needle, kind) {
			out = append(out, projectCandidate{ID: l.ID, Title: l.Title})
		}
	}
	return out
}

func parseTaskDate(raw string) (model.NullableTime, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return model.NullableTime{}, nil
	}

	parsed, err := dateparse.ParseDateExpr(time.Now(), value)
	if err != nil {
		return model.NullableTime{}, err
	}

	return model.NewNullableTime(parsed), nil
}

func parseDeferDuration(raw string) (time.Duration, error) {
	return dateparse.ParseDurationExpr(raw)
}

func reminderFromFlag(raw string, due model.NullableTime) ([]model.TaskReminder, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return []model.TaskReminder{}, nil
	}

	if strings.EqualFold(value, "due") {
		if !due.Valid {
			return nil, fmt.Errorf("--reminder=due requires a due date")
		}

		return []model.TaskReminder{{Reminder: due}}, nil
	}

	reminder, err := parseTaskDate(value)
	if err != nil {
		return nil, err
	}

	if !reminder.Valid {
		return []model.TaskReminder{}, nil
	}

	return []model.TaskReminder{{Reminder: reminder}}, nil
}

func taskLabelIDs(task model.Task) []int64 {
	if len(task.Labels) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(task.Labels))
	for _, label := range task.Labels {
		if label.ID > 0 {
			ids = append(ids, label.ID)
		}
	}

	return ids
}

func toggleTaskLabelIDs(current []int64, toggles []int64) []int64 {
	if len(toggles) == 0 {
		return append([]int64(nil), current...)
	}

	ids := make([]int64, 0, len(current))
	present := make(map[int64]struct{}, len(current))
	for _, id := range current {
		if id <= 0 {
			continue
		}
		if _, exists := present[id]; exists {
			continue
		}
		present[id] = struct{}{}
		ids = append(ids, id)
	}

	for _, id := range toggles {
		if id <= 0 {
			continue
		}

		if _, exists := present[id]; exists {
			delete(present, id)
			for i := range ids {
				if ids[i] != id {
					continue
				}
				ids = append(ids[:i], ids[i+1:]...)
				break
			}
			continue
		}

		present[id] = struct{}{}
		ids = append(ids, id)
	}

	return ids
}

func writeTaskOutput(cmd *cobra.Command, tasks []model.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	if flagJSON {
		if len(tasks) == 1 {
			return output.WriteJSONSingle(cmd.OutOrStdout(), tasks[0])
		}
		return output.WriteJSONList(cmd.OutOrStdout(), tasks)
	}

	now := time.Now()
	for i, task := range tasks {
		if i > 0 {
			if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
				return err
			}
		}

		text := output.FormatKeyValuesOmitEmpty([]output.KeyValue{
			{Key: "ID", Value: strconv.FormatInt(task.ID, 10)},
			{Key: "Title", Value: task.Title},
			{Key: "Done", Value: formatBoolText(task.Done)},
			{Key: "Due", Value: output.FormatDateText(task.DueDate, now)},
			{Key: "Project", Value: strconv.FormatInt(task.ProjectID, 10)},
			{Key: "Priority", Value: strconv.Itoa(task.Priority)},
			{Key: "Favorite", Value: formatBoolText(task.IsFavorite)},
		})

		if _, err := fmt.Fprintln(cmd.OutOrStdout(), text); err != nil {
			return err
		}
	}

	return nil
}

// formatBoolText renders booleans for the compact write-back view: "yes"/""
// so that the false case is omitted by FormatKeyValuesOmitEmpty.
func formatBoolText(value bool) string {
	if value {
		return "yes"
	}
	return ""
}

func shiftFromNowIfPast(current model.NullableTime, now time.Time, delta time.Duration) model.NullableTime {
	if !current.Valid {
		return current
	}

	base := current.Time
	if base.Before(now) {
		base = now
	}

	return model.NewNullableTime(base.Add(delta))
}
