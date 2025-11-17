package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Template represents an X.509 or SSH template stored via the admin API.
type Template struct {
	Name     string            `json:"name"`
	Body     string            `json:"body"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CreateTemplate stores a template that can later be attached to a provisioner.
func (c *Client) CreateTemplate(ctx context.Context, tmpl Template) error {
	return c.templateMutation(ctx, http.MethodPost, fmt.Sprintf("%s/admin/templates", c.baseURL), tmpl)
}

// UpdateTemplate updates the rendered template payload in-place.
func (c *Client) UpdateTemplate(ctx context.Context, tmpl Template) error {
	path := fmt.Sprintf("%s/admin/templates/%s", c.baseURL, tmpl.Name)
	return c.templateMutation(ctx, http.MethodPut, path, tmpl)
}

// DeleteTemplate removes a stored template by name. Missing templates are ignored.
func (c *Client) DeleteTemplate(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/admin/templates/%s", c.baseURL, name), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.adminToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

// GetTemplate fetches a template definition by name.
func (c *Client) GetTemplate(ctx context.Context, name string) (*Template, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/admin/templates/%s", c.baseURL, name), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.adminToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var tmpl Template
	if err := json.NewDecoder(resp.Body).Decode(&tmpl); err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (c *Client) templateMutation(ctx context.Context, method, url string, tmpl Template) error {
	b, err := json.Marshal(tmpl)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}
