package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Admin represents an admin user configuration.
type Admin struct {
	Name        string `json:"name"`
	Provisioner string `json:"provisioner"`
}

// CreateAdmin adds a new admin using the admin API.
func (c *Client) CreateAdmin(ctx context.Context, a Admin) error {
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/admin/admins", c.baseURL), bytes.NewReader(b))
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
	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

// ReplaceAdmin updates an existing admin entry.
func (c *Client) ReplaceAdmin(ctx context.Context, currentName, currentProvisioner string, a Admin) error {
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("%s/admin/admins/%s?provisioner=%s", c.baseURL, currentName, currentProvisioner)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, path, bytes.NewReader(b))
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
	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

// DeleteAdmin removes an admin via the admin API.
func (c *Client) DeleteAdmin(ctx context.Context, name, provisioner string) error {
	path := fmt.Sprintf("%s/admin/admins/%s?provisioner=%s", c.baseURL, name, provisioner)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, path, nil)
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

// GetAdmin retrieves an admin by name and provisioner.
func (c *Client) GetAdmin(ctx context.Context, name, provisioner string) (*Admin, error) {
	path := fmt.Sprintf("%s/admin/admins/%s?provisioner=%s", c.baseURL, name, provisioner)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
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
	var out Admin
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
