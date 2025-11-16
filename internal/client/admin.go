package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Admin struct {
	Name        string `json:"name"`
	Provisioner string `json:"provisioner"`
}

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
	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func (c *Client) UpdateAdmin(ctx context.Context, name, provisioner string, a Admin) error {
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("%s/admin/admins/%s?provisioner=%s", c.baseURL, name, provisioner)
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
	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

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
	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

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
	if resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var out Admin
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
