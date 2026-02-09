# ==============================================================================
# Namespace Operator - Makefile
# Toolchain: Go + BuildKit (containerd)
# ==============================================================================

# ------------------------------------------------------------------------------
# OS / ARCH detection (MUST be first)
# ------------------------------------------------------------------------------

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
  OS := linux
endif
ifeq ($(UNAME_S),Darwin)
  OS := darwin
endif

ifeq ($(UNAME_M),x86_64)
  ARCH := amd64
endif
ifeq ($(UNAME_M),arm64)
  ARCH := arm64
endif
ifeq ($(UNAME_M),aarch64)
  ARCH := arm64
endif

# ------------------------------------------------------------------------------
# General
# ------------------------------------------------------------------------------
BIN_DIR := $(CURDIR)/.bin
ENVTEST := $(BIN_DIR)/setup-envtest

# ------------------------------------------------------------------------------
# Project
# ------------------------------------------------------------------------------
PROJECT_NAME := namespace-operator
VERSION ?= $(shell git describe --tags --abbrev=0 | sed 's/^v//')
VERSION := 1.2.0
IMG ?= baabdoul/namespace-operator:$(VERSION)

# ------------------------------------------------------------------------------
# Helm
# ------------------------------------------------------------------------------
HELM_CHART := manifests/charts/namespace-operator
HELM_RELEASE := namespace-operator
HELM_NAMESPACE := namespace-operator-system


# ------------------------------------------------------------------------------
# Tools
# ------------------------------------------------------------------------------
GO ?= go
HELM ?= helm
KUBECTL ?= kubectl
CONTROLLER_GEN ?= controller-gen

# ------------------------------------------------------------------------------
# Go (vendored, versioned + symlink)
# ------------------------------------------------------------------------------
GO_VERSION := 1.24.13
GO_VERSION := 1.22.7

GO_TGZ := go$(GO_VERSION).$(OS)-$(ARCH).tar.gz
GO_URL := https://go.dev/dl/$(GO_TGZ)

GO_BASE_DIR := $(BIN_DIR)
GO_VERSION_DIR := $(GO_BASE_DIR)/go-$(GO_VERSION)
GO_SYMLINK := $(GO_BASE_DIR)/go

GO_BIN := $(GO_SYMLINK)/bin/go
GO := $(GO_BIN)


# Disable CGO (no system linker required)
export CGO_ENABLED := 0
export GOOS := $(OS)
export GOARCH := $(ARCH)
export PATH := $(BIN_DIR)/go/bin:$(BIN_DIR):$(PATH)
# ------------------------------------------------------------------------------
# BuildKit / buildctl (vendored)
# ------------------------------------------------------------------------------

BUILDKIT_VERSION := v0.27.1
BUILDKIT_TGZ := buildkit-$(BUILDKIT_VERSION).$(OS)-$(ARCH).tar.gz
BUILDKIT_URL := https://github.com/moby/buildkit/releases/download/$(BUILDKIT_VERSION)/$(BUILDKIT_TGZ)

BUILDCTL := $(BIN_DIR)/buildctl

# ------------------------------------------------------------------------------
# Kubernetes tooling (vendored)
# ------------------------------------------------------------------------------

CONTROLLER_GEN := $(BIN_DIR)/controller-gen
KUSTOMIZE := $(BIN_DIR)/kustomize

# ------------------------------------------------------------------------------
# Default target
# ------------------------------------------------------------------------------

all: build

# ------------------------------------------------------------------------------
# Toolchain installation
# ------------------------------------------------------------------------------

.PHONY: go
go: ## Install Go locally (versioned) and update symlink
	@mkdir -p $(BIN_DIR)
	@if [ -x "$(GO_VERSION_DIR)/bin/go" ]; then \
		echo "✅ Go $(GO_VERSION) already installed"; \
	else \
		echo "⬇️  Installing Go $(GO_VERSION) ($(OS)/$(ARCH))"; \
		echo "    → $(GO_URL)"; \
		curl -fL $(GO_URL) -o /tmp/$(GO_TGZ); \
		tar -xzf /tmp/$(GO_TGZ) -C /tmp; \
		mv /tmp/go $(GO_VERSION_DIR); \
		rm -f /tmp/$(GO_TGZ); \
		echo "✅ Go $(GO_VERSION) installed"; \
	fi
	@ln -sfn go-$(GO_VERSION) $(GO_SYMLINK)
	@echo "🔗 go -> go-$(GO_VERSION)"
	@$(GO_BIN) version

.PHONY: buildctl
buildctl: ## Install buildctl locally if missing
	@mkdir -p $(BIN_DIR)
	@if [ -x "$(BUILDCTL)" ]; then \
		echo "✅ buildctl already installed"; \
	else \
		echo "⬇️  Downloading buildctl $(BUILDKIT_VERSION) ($(OS)/$(ARCH))"; \
		echo "    → $(BUILDKIT_URL)"; \
		curl -fL $(BUILDKIT_URL) -o /tmp/$(BUILDKIT_TGZ); \
		tar -xzf /tmp/$(BUILDKIT_TGZ) -C /tmp; \
		mv /tmp/bin/buildctl $(BUILDCTL); \
		chmod +x $(BUILDCTL); \
		rm -rf /tmp/$(BUILDKIT_TGZ) /tmp/bin; \
		echo "✅ buildctl installed"; \
	fi

.PHONY: controller-gen
controller-gen: go ## Install controller-gen locally if missing
	@mkdir -p $(BIN_DIR)
	@if [ ! -x "$(CONTROLLER_GEN)" ]; then \
		echo "⬇️  Installing controller-gen"; \
		GOBIN=$(BIN_DIR) $(GO) install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0 ;\
	fi

.PHONY: kustomize
kustomize: go ## Install kustomize locally if missing
	@mkdir -p $(BIN_DIR)
	@if [ ! -x "$(KUSTOMIZE)" ]; then \
		echo "⬇️  Installing kustomize"; \
		GOBIN=$(BIN_DIR) $(GO) install sigs.k8s.io/kustomize/kustomize/v5@latest ;\
	fi

.PHONY: envtest
envtest: go ## Install envtest locally if missing
	@mkdir -p $(BIN_DIR)
	@if [ ! -x "$(ENVTEST)" ]; then \
		echo "⬇️  Installing envtest"; \
		GOBIN=$(BIN_DIR) $(GO) install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest ;\
	fi

# ------------------------------------------------------------------------------
# Development
# ------------------------------------------------------------------------------

.PHONY: generate
generate: controller-gen ## Generate deepcopy code
	$(CONTROLLER_GEN) object paths="./..."

.PHONY: manifests
manifests: controller-gen ## Generate CRDs and RBAC manifests
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd paths="./..." \
		output:crd:artifacts:config=manifests/charts/namespace-operator/crds

.PHONY: fmt
fmt: go ## Run go fmt
	$(GO) fmt ./...

.PHONY: vet
vet: go ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: vet
	$(GO) fmt ./...

.PHONY: test
test: envtest generate manifests fmt vet ## Run controller tests
	$(ENVTEST) use latest --bin-dir $(BIN_DIR) >/dev/null
	KUBEBUILDER_ASSETS=$$($(ENVTEST) use latest --bin-dir $(BIN_DIR) -p path) \
	$(GO) test ./... -coverprofile cover.out

.PHONY: crd
crd:
	$(CONTROLLER_GEN) crd paths=./api/... \
	  output:crd:artifacts:config=manifests/charts/namespace-operator/crds

# ------------------------------------------------------------------------------
# Build
# ------------------------------------------------------------------------------

.PHONY: build
build: fmt vet ## Build manager binary
	$(GO) build -o bin/manager main.go

.PHONY: run
run: generate manifests fmt vet ## Run controller locally
	$(GO) run ./main.go

# ------------------------------------------------------------------------------
# Container image (BuildKit / containerd)
# ------------------------------------------------------------------------------

.PHONY: image-build
image-build: buildctl ## Build OCI image (no Docker)
	sudo $(BUILDCTL) build \
	  --frontend dockerfile.v0 \
	  --local context=. \
	  --local dockerfile=. \
	  --output type=image,name=$(IMG),push=false

.PHONY: image-push
image-push: buildctl ## Build and push OCI image
	sudo $(BUILDCTL) build \
	  --frontend dockerfile.v0 \
	  --local context=. \
	  --local dockerfile=. \
	  --output type=image,name=$(IMG),push=true

# ------------------------------------------------------------------------------
# Kubernetes
# ------------------------------------------------------------------------------
.PHONY: helm-install
helm-install:
	$(HELM) install $(HELM_RELEASE) $(HELM_CHART) \
	  --namespace $(HELM_NAMESPACE) \
	  --create-namespace

.PHONY: helm-upgrade
helm-upgrade:
	$(HELM) upgrade $(HELM_RELEASE) $(HELM_CHART) \
	  --namespace $(HELM_NAMESPACE)

.PHONY: helm-uninstall
helm-uninstall:
	$(HELM) uninstall $(HELM_RELEASE) \
	  --namespace $(HELM_NAMESPACE)

.PHONY: helm-template
helm-template:
	$(HELM) template $(HELM_RELEASE) $(HELM_CHART) \
	  --namespace $(HELM_NAMESPACE)

.PHONY: helm-lint
helm-lint:
	$(HELM) lint $(HELM_CHART)

.PHONY: crd-apply
crd-apply: ## Apply CRDs to the cluster
	kubectl apply -f manifests/charts/namespace-operator/crds

# ------------------------------------------------------------------------------
# Info / Debug
# ------------------------------------------------------------------------------

.PHONY: info
info: ## Display detected environment
	@echo "OS        : $(OS)"
	@echo "ARCH      : $(ARCH)"
	@echo "Go        : $(GO_VERSION)"
	@echo "BuildKit  : $(BUILDKIT_VERSION)"
	@echo "BIN_DIR   : $(BIN_DIR)"

# ------------------------------------------------------------------------------
# Help
# ------------------------------------------------------------------------------

.PHONY: help
help: ## Show help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} \
	/^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2 }' \
	$(MAKEFILE_LIST)