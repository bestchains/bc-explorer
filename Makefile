GOOS ?= linux
GOARCH ?= $(shell go env GOARCH)
GOFLAGS ?=""
SOURCES := $(shell find . -type f  -name '*.go')

REGISTRY ?= "docker.io/hyperledgerk8s"

DOCKER_TARGETS := bc-explorer

TARGETS := ${DOCKER_TARGETS}

WHAT ?= $(TARGETS)

# Build binary
#
# Args:
#   WHAT:   Target to build.
#   GOOS:   OS to build.
#   GOARCH: Arch to build.
#
# Example:
#   make all

.PHONY: all
all: binary

binary:
	@GOFLAGS=$(GOFLAGS) BUILD_PLATFORMS=$(GOOS)/$(GOARCH) hack/build.sh $(WHAT)

.PHONY: clean
clean:
	rm -rf _output

.PHONY: test
test:
	@hack/run-unit-tests.sh

.PHONY: verify
verify:
	@hack/verify-copyright.sh
	@hack/verify-golangci-lint.sh
	@hack/verify-shfmt.sh

.PHONY: golanglint-fix
golanglint-fix:
	@hack/fix-golang-lint-error.sh

# Build image.
#
# Args:
#   WHAT:        Target to build.
#   GOOS:        OS to build.
#   GOARCH:      Arch to build.
#   OUTPUT_TYPE: Destination to save image(docker/registry).
#
# Example:
#   make image
.PHONY: images
image:
	@REGISTRY=$(REGISTRY) OUTPUT_TYPE=$(OUTPUT_TYPE) BUILD_PLATFORMS=$(GOOS)/$(GOARCH) hack/build-image.sh $(filter ${DOCKER_TARGETS}, ${WHAT})
