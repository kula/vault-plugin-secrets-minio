PROJECT		= "github.com/kula/vault-plugin-secrets-minio"
GOFILES		= $(shell find . -name "*.go")

default: vault-plugin-secrets-minio

vault-plugin-secrets-minio: $(GOFILES)
	go build ./cmd/vault-plugin-secrets-minio

clean:
	rm -f vault-plugin-secrets-minio

test: vault-plugin-secrets-minio
	/bin/bash test/test.sh

deps:
	go get ./...
	# If you don't do this, you get a panic because /debug/requests
	# is registered twice, because both minio and vault vendor
	# golang.org/x/net/trace/ but neither seem to use it?
	# https://github.com/etcd-io/etcd/issues/9357
	rm -f ${GOPATH}/src/github.com/minio/minio/vendor/golang.org/x/net/trace/
	rm -f ${GOPATH}/src/github.com/hashicorp/vault/vendor/golang.org/x/net/trace/

.PHONY: default clean test deps
