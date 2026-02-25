package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/model"
)

func (c *Client) GetLabels(ctx context.Context) ([]model.Label, error) {
	var labels []model.Label
	if err := c.GetJSON(ctx, "/labels", &labels); err != nil {
		return nil, err
	}

	return labels, nil
}

func (c *Client) CreateLabel(ctx context.Context, title string) (model.Label, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return model.Label{}, fmt.Errorf("title is required")
	}

	payload := model.LabelCreatePayload{Title: title}

	var label model.Label
	if err := c.PutJSON(ctx, "/labels", payload, &label); err != nil {
		return model.Label{}, err
	}

	return label, nil
}
