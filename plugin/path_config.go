package minio

import (
    "context"
    "fmt"

    "github.com/hashicorp/errwrap"
    "github.com/hashicorp/vault/logical"
    "github.com/hashicorp/vault/logical/framework"
)

// Define the CRU functions for the config path
func (b *backend) pathConfigCRUD() *framework.Path {
    return &framework.Path{
	Pattern: fmt.Sprintf("config/?$"),
	HelpSynopsis: "Configure the Minio connection.",
	HelpDescription: "Use this endpoint to set the Minio endpoint, accessKeyId, secretAccessKey and SSL settings.",

	Fields: map[string]*framework.FieldSchema{
	    "endpoint": &framework.FieldSchema{
		Type: framework.TypeString,
		Description: "The Minio server endpoint.",
	    },
	    "accessKeyId": &framework.FieldSchema{
		Type: framework.TypeString,
		Description: "The Minio administrative key ID.",
	    },
	    "secretAccessKey": &framework.FieldSchema{
		Type: framework.TypeString,
		Description: "The Minio administrative secret access key.",
	    },
	    "useSSL": &framework.FieldSchema{
		Type: framework.TypeBool,
		Description: "(Optional, default `false`) Use SSL to connect to the Minio server.",
	    },
	},

	Callbacks: map[logical.Operation]framework.OperationFunc{
	    logical.ReadOperation: b.pathConfigRead,
	    logical.UpdateOperation: b.pathConfigUpdate,
	},
    }
}

// Read the current configuration
func (b *backend) pathConfigRead(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
    c, err := b.GetConfig(ctx, req.Storage);
    if err != nil {
	return nil, err
    }

    return &logical.Response{
	Data: map[string]interface{}{
	    "endpoint": c.Endpoint,
	    "accessKeyId": c.AccessKeyId,
	    "secretAccessKey": c.SecretAccessKey,
	    "useSSL": c.UseSSL,
	},
    }, nil
}

// Update the configuration
func (b *backend) pathConfigUpdate(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
    c, err := b.GetConfig(ctx, req.Storage);
    if err != nil {
	return nil, err
    }

    // Update the internal configuration
    changed, err := c.Update(d)
    if err != nil {
	return nil, logical.CodedError(400, err.Error())
    }

    // If we changed the configuration, store it
    if changed {
	// Make a new storage entry
	entry, err := logical.StorageEntryJSON("config", c)
	if err != nil {
	    return nil, errwrap.Wrapf("failed to generate JSON configuration: {{err}}", err)
	}

	// And store it
	if err := req.Storage.Put(ctx, entry); err != nil {
	    return nil, errwrap.Wrapf("Failed to persist configuration: {{err}}", err)
	}

    }

    // Destroy any old client which may exist so we get a new one
    // with the next request
    b.invalidateMadminClient()

    return nil, nil
}
