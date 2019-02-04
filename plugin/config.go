package minio

import (
    "context"
    "strings"

    "github.com/hashicorp/errwrap"
    "github.com/hashicorp/vault/logical"
    "github.com/hashicorp/vault/logical/framework"
)

// The stored plugin configuration

type Config struct {
    Endpoint string `json:"endpoint"`
    AccessKeyId string `json:"accessKeyId"`
    SecretAccessKey string `json:"secretAccessKey"`
    UseSSL bool `json:"useSSL"`
    Configured bool `json:"is_configured"`
}

// A default, empty configuration

func DefaultConfig() *Config {
    return &Config{
	Endpoint: "",
	AccessKeyId: "",
	SecretAccessKey: "",
	UseSSL: false,
	Configured: false,
    }
}

// Update the configuration with new values

func (c *Config) Update(d *framework.FieldData) (bool, error) {
    if d == nil {
	return false, nil
    }

    changed := false

    keys := []string{"endpoint", "accessKeyId", "secretAccessKey"}

    for _, key := range keys {
	if v, ok := d.GetOk(key); ok {
	    nv := strings.TrimSpace(v.(string))

	    switch key {
	    case "endpoint":
		c.Endpoint = nv
		c.Configured = true
		changed = true
	    case "accessKeyId":
		c.AccessKeyId = nv
		c.Configured = true
		changed = true
	    case "secretAccessKey":
		c.SecretAccessKey = nv
		c.Configured = true
		changed = true
	    }
	}
    }

    if v, ok := d.GetOk("useSSL"); ok {
	nv := v.(bool)
	c.UseSSL = nv
	c.Configured = true
	changed = true
    }

    return changed, nil
}


// Retrieve the configuration from the backend storage. In the event
// the plugin has not been configured, a default, empty configuration
// is returned

func (b *backend) GetConfig(ctx context.Context, s logical.Storage) (*Config, error) {
    c := DefaultConfig()

    entry, err := s.Get(ctx, "config");
    if err != nil {
	return nil, errwrap.Wrapf("failed to get configuration from backend: {{err}}", err)
    }

    if entry == nil || len(entry.Value) == 0 {
	return c, nil
    }

    if err := entry.DecodeJSON(&c); err != nil {
	return nil, errwrap.Wrapf("failed to decode configuration: {{err}}", err)
    }

    return c, nil
}
