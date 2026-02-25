package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

var nullableTimeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// NullableTime supports null, empty string, and common date layouts from API payloads.
type NullableTime struct {
	Time  time.Time
	Valid bool
}

func NewNullableTime(t time.Time) NullableTime {
	return NullableTime{Time: t, Valid: true}
}

func (nt *NullableTime) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)

	if bytes.Equal(trimmed, []byte("null")) {
		*nt = NullableTime{}
		return nil
	}

	var value string
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return fmt.Errorf("nullable time must be string or null: %w", err)
	}

	if value == "" {
		*nt = NullableTime{}
		return nil
	}

	for _, layout := range nullableTimeLayouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			*nt = NullableTime{Time: parsed, Valid: true}
			return nil
		}
	}

	return fmt.Errorf("invalid time value %q", value)
}

func (nt NullableTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(nt.Time.Format(time.RFC3339Nano))
}

func (nt NullableTime) IsZero() bool {
	return !nt.Valid
}

func (nt NullableTime) TimeOrNil() *time.Time {
	if !nt.Valid {
		return nil
	}

	t := nt.Time
	return &t
}

type TaskReminder struct {
	ID             int64        `json:"id,omitempty"`
	Reminder       NullableTime `json:"reminder,omitempty"`
	RelativePeriod int64        `json:"relative_period,omitempty"`
	RelativeTo     string       `json:"relative_to,omitempty"`
}

type Task struct {
	ID          int64          `json:"id,omitempty"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	Done        bool           `json:"done,omitempty"`
	DoneAt      NullableTime   `json:"done_at,omitempty"`
	DueDate     NullableTime   `json:"due_date,omitempty"`
	Reminders   []TaskReminder `json:"reminders,omitempty"`
	ProjectID   int64          `json:"project_id,omitempty"`
	Priority    int            `json:"priority,omitempty"`
	Labels      []Label        `json:"labels,omitempty"`
	IsFavorite  bool           `json:"is_favorite,omitempty"`
	Created     NullableTime   `json:"created,omitempty"`
	Updated     NullableTime   `json:"updated,omitempty"`
}

type TaskCreatePayload struct {
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Done        bool           `json:"done,omitempty"`
	DueDate     *NullableTime  `json:"due_date,omitempty"`
	Reminders   []TaskReminder `json:"reminders,omitempty"`
	ProjectID   *int64         `json:"project_id,omitempty"`
	Priority    *int           `json:"priority,omitempty"`
	LabelIDs    []int64        `json:"label_ids,omitempty"`
	IsFavorite  *bool          `json:"is_favorite,omitempty"`
}

type TaskUpdatePayload struct {
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	Done        *bool           `json:"done,omitempty"`
	DueDate     *NullableTime   `json:"due_date,omitempty"`
	DoneAt      *NullableTime   `json:"done_at,omitempty"`
	Reminders   *[]TaskReminder `json:"reminders,omitempty"`
	ProjectID   *int64          `json:"project_id,omitempty"`
	Priority    *int            `json:"priority,omitempty"`
	LabelIDs    *[]int64        `json:"label_ids,omitempty"`
	IsFavorite  *bool           `json:"is_favorite,omitempty"`
}
