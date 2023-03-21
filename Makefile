GOOS ?= linux
GOARCH ?= $(shell go env GOARCH)
GOFLAGS ?=""
SOURCES := $(shell find . -type f  -name '*.go')

.PHONY: verify
verify:
	@hack/verify-copyright.sh
