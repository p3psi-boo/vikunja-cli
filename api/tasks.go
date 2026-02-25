package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/model"
)

type TaskFilter struct {
	All      bool
	Project  string
	Labels   []string
	Priority string
	Due      string
	Favorite bool
	Filters  []string
	Sort     string
	Limit    int
}

func (c *Client) GetTasks(ctx context.Context, filter TaskFilter) ([]model.Task, error) {
	values := url.Values{}

	if !filter.All {
		values.Set("done", "false")
	}

	if project := strings.TrimSpace(filter.Project); project != "" {
		values.Set("project", project)
	}

	for _, label := range filter.Labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		values.Add("label", label)
	}

	if priority := strings.TrimSpace(filter.Priority); priority != "" {
		values.Set("priority", priority)
	}

	if due := strings.TrimSpace(filter.Due); due != "" {
		values.Set("due", due)
	}

	if filter.Favorite {
		values.Set("favorite", "true")
	}

	for _, rawFilter := range filter.Filters {
		rawFilter = strings.TrimSpace(rawFilter)
		if rawFilter == "" {
			continue
		}
		values.Add("filter", rawFilter)
	}

	if sort := strings.TrimSpace(filter.Sort); sort != "" {
		values.Set("sort", sort)
	}

	if filter.Limit > 0 {
		values.Set("limit", strconv.Itoa(filter.Limit))
	}

	path := "/tasks"
	query := values.Encode()
	if query != "" {
		path += "?" + query
	}

	var tasks []model.Task
	if err := c.GetJSON(ctx, path, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (c *Client) GetTask(ctx context.Context, id int64) (model.Task, error) {
	var task model.Task
	if err := c.GetJSON(ctx, fmt.Sprintf("/tasks/%d", id), &task); err != nil {
		return model.Task{}, err
	}

	return task, nil
}

func (c *Client) CreateTask(ctx context.Context, payload model.TaskCreatePayload) (model.Task, error) {
	if payload.ProjectID == nil || *payload.ProjectID <= 0 {
		return model.Task{}, fmt.Errorf("project id must be a positive integer")
	}

	if strings.TrimSpace(payload.Title) == "" {
		return model.Task{}, fmt.Errorf("task title is required")
	}

	var task model.Task
	path := fmt.Sprintf("/projects/%d/tasks", *payload.ProjectID)
	if err := c.PutJSON(ctx, path, payload, &task); err != nil {
		return model.Task{}, err
	}

	return task, nil
}

func (c *Client) UpdateTask(ctx context.Context, id int64, payload model.TaskUpdatePayload) (model.Task, error) {
	if id <= 0 {
		return model.Task{}, fmt.Errorf("task id must be a positive integer")
	}

	var task model.Task
	if err := c.PostJSON(ctx, fmt.Sprintf("/tasks/%d", id), payload, &task); err != nil {
		return model.Task{}, err
	}

	return task, nil
}

func (c *Client) DeleteTask(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("task id must be a positive integer")
	}

	return c.DeleteJSON(ctx, fmt.Sprintf("/tasks/%d", id), nil, nil)
}
