package minio

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/hashicorp/errwrap"
    "github.com/hashicorp/vault/logical"
)

var (
    ErrRoleNotFound = errors.New("role not found")
)

// A role stored in the storage backend
type Role struct {

    // Policy is the policy the created user will have on the
    // Minio server
    Policy string `json:"policy"`

    // UserNamePrefix is what we prepend to the key name when we
    // create it, followed by the Vault request ID which asked
    // for the key to be made

    UserNamePrefix string `json:"user_name_prefix"`

    // DefaultTTL is the TTL which will be applied to keys if no
    // TTL is requested
    DefaultTTL time.Duration `json:"default_ttl"`

    // MaxTTL is the maximum any TTL can be for this role
    MaxTTL time.Duration `json:"max_ttl"`
}

// List Roles

func (b* backend) ListRoles(ctx context.Context, s logical.Storage) ([]string, error) {
    roles, err := s.List(ctx, "roles/")
    if err != nil {
	return nil, errwrap.Wrapf("Unable to retrieve list of roles: {{err}}", err)
    }

    return roles, nil
}

// Get Role

func (b* backend) GetRole(ctx context.Context, s logical.Storage, role string) (*Role, error) {
    r, err := s.Get(ctx, "roles/"+role)
    if err != nil { 
	return nil, errwrap.Wrapf(fmt.Sprintf("Unable to retrieve role %q: {{err}}", role), err)
    }

    if r == nil {
	return nil, ErrRoleNotFound
    }

    var rv Role
    if err := r.DecodeJSON(&rv); err != nil {
	return nil, errwrap.Wrapf(fmt.Sprintf("Unable to decode role %q: {{err}}", role), err)
    }

    return &rv, nil
}

