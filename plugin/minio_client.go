package minio

import (
    "context"
    "errors"

    "github.com/hashicorp/vault/logical"

    "github.com/minio/minio/pkg/madmin"
)

// Convenience function to get a new madmin client
func (b *backend) getMadminClient(ctx context.Context, s logical.Storage) (*madmin.AdminClient, error) {

    b.Logger().Debug("getMadminClient, getting clientMutext.RLock")
    b.clientMutex.Lock()
    defer b.clientMutex.Unlock()

    if b.client != nil {
	b.Logger().Debug("Already have client, returning")
	return b.client, nil
    }

    // Don't have client, look up configuration and gin up new client
    b.Logger().Info("getMadminClient, need new client and looking up config")

    c, err := b.GetConfig(ctx, s)
    if err != nil {
	b.Logger().Error("Error fetching config in getMadminClient", "error", err)
	return nil, err
    }

    if c.Endpoint == "" {
	err = errors.New("Endpoint not set when trying to create new madmin client")
	b.Logger().Error("Error", "error", err)
	return nil, err
    }

    if c.AccessKeyId == "" {
	err = errors.New("AccessKeyId not set when trying to create new madmin client")
	b.Logger().Error("Error", "error", err)
	return nil, err
    }

    if c.SecretAccessKey == "" {
	err = errors.New("SecretAccessKey not set when trying to create new madmin client")
	b.Logger().Error("Error", "error", err)
	return nil, err
    }

    client, err := madmin.New(c.Endpoint, c.AccessKeyId, c.SecretAccessKey, c.UseSSL)
    if err != nil {
	b.Logger().Error("Error getting new madmin client", "error", err)
	return nil, err
    }
    
    b.client = client
    return b.client, nil
}

// Call this to invalidate the current backend client
func (b *backend) invalidateMadminClient() {
    b.Logger().Debug("invalidateMadminClient")
    
    b.clientMutex.Lock()
    defer b.clientMutex.Unlock()

    b.client = nil
}
