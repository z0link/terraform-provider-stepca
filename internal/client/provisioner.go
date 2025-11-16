package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Provisioner struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Admin bool   `json:"admin,omitempty"`
}

func (c *Client) CreateProvisioner(ctx context.Context, p Provisioner) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/admin/provisioners", c.baseURL), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.adminToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func (c *Client) UpdateProvisioner(ctx context.Context, name string, p Provisioner) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/admin/provisioners/%s", c.baseURL, name), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.adminToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func (c *Client) DeleteProvisioner(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/admin/provisioners/%s", c.baseURL, name), nil)
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
	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func (c *Client) GetProvisioner(ctx context.Context, name string) (*Provisioner, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/admin/provisioners/%s", c.baseURL, name), nil)
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
	if resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var out Provisioner
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
