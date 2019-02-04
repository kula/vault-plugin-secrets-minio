package minio

import (
    "context"
    "fmt"
    "time"

    "github.com/hashicorp/errwrap"
    "github.com/hashicorp/vault/logical"
    "github.com/hashicorp/vault/logical/framework"
)

// Define the R functions for the keys path
func (b *backend) pathKeysRead() *framework.Path {
    return &framework.Path{
	Pattern: fmt.Sprintf("keys/" + framework.GenericNameRegex("role")),
	HelpSynopsis: "Provision a key for this role.",

	Fields: map[string]*framework.FieldSchema{
	    "role": &framework.FieldSchema{
		Type: framework.TypeString,
		Description: "Name of role.",
	    },
	    "ttl": &framework.FieldSchema{
		Type: framework.TypeDurationSecond,
		Description: "Lifetime of accessKey in seconds.",
	    },
	},

	Callbacks: map[logical.Operation]framework.OperationFunc{
	    logical.ReadOperation: b.pathKeyRead,
	},
    }
}


// Read a new key
func (b *backend) pathKeyRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

    roleName := d.Get("role").(string)

    role, err := b.GetRole(ctx, req.Storage, roleName)
    if err != nil {
	return nil, errwrap.Wrapf("error fetching role: {{err}}", err)
    }

    newKeyName := fmt.Sprintf("%s%s", role.UserNamePrefix, req.ID)

    // Calculate lifetime
    reqTTL := time.Duration(d.Get("ttl").(int)) * time.Second
    ttl, _, err := framework.CalculateTTL(b.System(), 0, role.DefaultTTL, reqTTL, role.MaxTTL, 0, time.Time{})
    if err != nil {
	return nil, err
    }

    // Generate new accessKey
    newKey, err := b.minioAccessKeyCreate(ctx, req.Storage, newKeyName, role.Policy)
    if err != nil {
	return nil, err
    }

    // Gin up response
    resp := b.Secret(minioKeyType).Response(map[string]interface{}{
	// Returned secret
	"accessKeyId": newKeyName,
	"secretAccessKey": newKey.SecretKey,
	"accountStatus": newKey.Status,
	"policy": newKey.PolicyName,
    }, map[string]interface{}{
	// Internal Data
	"accessKeyId": newKeyName,
    })

    resp.Secret.TTL = ttl
    resp.Secret.MaxTTL = role.MaxTTL

    return resp, nil
}
