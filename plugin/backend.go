package minio

import (
    "context"
    "sync"

    "github.com/hashicorp/vault/logical"
    "github.com/hashicorp/vault/logical/framework"

    "github.com/minio/minio/pkg/madmin"
)

type backend struct {
    *framework.Backend

    client *madmin.AdminClient

    // We're going to have to be able to rotate the client
    // if the mount configured credentials change, use
    // this to protect it
    clientMutex sync.RWMutex
}

// Factory returns a configured instance of the minio backend
func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
    b := Backend()
    if err := b.Setup(ctx, c); err != nil {
	return nil, err
    }

    b.Logger().Info("Plugin successfully initialized")
    return b, nil
}

// Backend returns a configured minio backend
func Backend() *backend {
    var b backend

    b.Backend = &framework.Backend{
	BackendType: logical.TypeLogical,
	Help: "The minio secrets backend provisions users on a Minio server",

	Paths: []*framework.Path{
	    // path_config.go
	    // ^config
	    b.pathConfigCRUD(),

	    // path_roles.go
	    // ^roles (LIST)
	    b.pathRoles(),
	    // ^roles/<role> 
	    b.pathRolesCRUD(),

	    // path_keys.go
	    // ^keys/<role>
	    b.pathKeysRead(),
	},

	Secrets: []*framework.Secret{
	    b.minioAccessKeys(),
	},
    }

    b.client = (*madmin.AdminClient)(nil)

    return &b
}
