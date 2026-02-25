package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/p3psi-boo/vikunja-cli/model"
)

func TestWriteJSONSingleTaskUsesRFC3339(t *testing.T) {
	due := time.Date(2026, 2, 5, 11, 22, 33, 987654321, time.UTC)
	task := model.Task{
		ID:      7,
		Title:   "ship",
		DueDate: model.NewNullableTime(due),
	}

	var buf bytes.Buffer
	if err := WriteJSONSingle(&buf, task); err != nil {
		t.Fatalf("WriteJSONSingle() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"due_date":"2026-02-05T11:22:33Z"`) {
		t.Fatalf("expected RFC3339 due_date in output, got %q", out)
	}
	if strings.Contains(out, ".987654321") {
		t.Fatalf("expected no nanoseconds in output, got %q", out)
	}
}

func TestFormatTaskTable(t *testing.T) {
	now := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)
	tasks := []model.Task{
		{
			ID:         1,
			Priority:   5,
			IsFavorite: true,
			Title:      "Pay rent",
			DueDate:    model.NewNullableTime(now.Add(48 * time.Hour)),
			ProjectID:  10,
			Labels: []model.Label{
				{Title: "home"},
				{Title: "finance"},
			},
		},
	}

	out := FormatTaskTable(tasks, map[int64]string{10: "Personal"}, now)
	if !strings.Contains(out, "ID") || !strings.Contains(out, "Priority") || !strings.Contains(out, "Star") {
		t.Fatalf("expected task table headers, got %q", out)
	}
	if !strings.Contains(out, "Pay rent") {
		t.Fatalf("expected task row, got %q", out)
	}
	if !strings.Contains(out, "in 2d") {
		t.Fatalf("expected relative due date, got %q", out)
	}
	if !strings.Contains(out, "finance,home") {
		t.Fatalf("expected labels list, got %q", out)
	}
}

func TestFormatKeyValues(t *testing.T) {
	out := FormatKeyValues([]KeyValue{{Key: "ID", Value: "5"}, {Key: "Title", Value: "Task"}})
	if !strings.Contains(out, "ID") || !strings.Contains(out, "Title") {
		t.Fatalf("expected key/value lines, got %q", out)
	}
}

func TestPrintInfoRespectsQuiet(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintInfo(&buf, true, "created %d", 7); err != nil {
		t.Fatalf("PrintInfo() error = %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output in quiet mode, got %q", buf.String())
	}
}
