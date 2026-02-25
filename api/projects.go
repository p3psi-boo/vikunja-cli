package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/model"
)

func (c *Client) GetProjects(ctx context.Context) ([]model.Project, error) {
	var projects []model.Project
	if err := c.GetJSON(ctx, "/projects", &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (c *Client) GetProject(ctx context.Context, id int64) (*model.Project, error) {
	if id <= 0 {
		return nil, fmt.Errorf("project id must be greater than 0")
	}

	var project model.Project
	if err := c.GetJSON(ctx, fmt.Sprintf("/projects/%d", id), &project); err != nil {
		return nil, err
	}

	return &project, nil
}

func (c *Client) CreateProject(ctx context.Context, payload model.ProjectCreatePayload) (*model.Project, error) {
	payload.Title = strings.TrimSpace(payload.Title)
	if payload.Title == "" {
		return nil, fmt.Errorf("project title is required")
	}

	var project model.Project
	if err := c.PutJSON(ctx, "/projects", payload, &project); err != nil {
		return nil, err
	}

	return &project, nil
}
