package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL          string
	token            string
	adminName        string
	adminKey         string
	adminProvisioner string
	adminToken       string
	httpClient       *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{baseURL: baseURL, token: token, httpClient: &http.Client{}}
}

func (c *Client) WithAdminToken(t string) *Client {
	c.adminToken = t
	return c
}

func (c *Client) WithAdminName(name string) *Client {
	c.adminName = name
	return c
}

func (c *Client) WithAdminKey(key string) *Client {
	c.adminKey = key
	return c
}

func (c *Client) WithAdminProvisioner(p string) *Client {
	c.adminProvisioner = p
	return c
}

// Sign sends a CSR to the /sign endpoint and returns the certificate PEM bytes.
func (c *Client) Sign(ctx context.Context, csr string) ([]byte, error) {
	body := map[string]string{"csr": csr, "ott": c.token}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/sign", c.baseURL), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var result struct {
		Cert string `json:"crt"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return []byte(result.Cert), nil
}

// Version retrieves the version string from the /version endpoint.
func (c *Client) Version(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/version", c.baseURL), nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(b)), nil
}

// RootCertificate retrieves the root certificate PEM from the /root endpoint.
func (c *Client) RootCertificate(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/root", c.baseURL), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Provisioner represents a simple provisioner configuration.
type Provisioner struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Admin bool   `json:"admin,omitempty"`
}

// CreateProvisioner adds a new provisioner using the admin API.
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
	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

// DeleteProvisioner removes a provisioner via the admin API.
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
	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

// GetProvisioner retrieves a provisioner by name.
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
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var out Provisioner
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

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
