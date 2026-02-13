# ==============================================================================
# Namespace Operator - Makefile (Multi-OS)
# ==============================================================================

# ------------------------------------------------------------------------------
# OS / ARCH detection
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
CONTROLLER_RUNTIME_VERSION := v0.22.5
ENVTEST_URL := https://github.com/kubernetes-sigs/controller-runtime/releases/download/$(CONTROLLER_RUNTIME_VERSION)/setup-envtest-$(OS)-$(ARCH)

K8S_VERSION := 1.29.0
# ------------------------------------------------------------------------------
# Project
# ------------------------------------------------------------------------------

PROJECT_NAME := namespace-operator
VERSION := 1.3.2
IMG ?= baabdoul/namespace-operator:$(VERSION)

# ------------------------------------------------------------------------------
# Helm
# ------------------------------------------------------------------------------

HELM_CHART := manifests/charts/namespace-operator
HELM_RELEASE := namespace-operator
HELM_NAMESPACE := namespace-operator-system
HELM_DOCS ?= helm-docs

# ------------------------------------------------------------------------------
# Go (vendored)
# ------------------------------------------------------------------------------

GO_VERSION := 1.24.0

GO_TGZ := go$(GO_VERSION).$(OS)-$(ARCH).tar.gz
GO_URL := https://go.dev/dl/$(GO_TGZ)

GO_BASE_DIR := $(BIN_DIR)
GO_VERSION_DIR := $(GO_BASE_DIR)/go-$(GO_VERSION)
GO_SYMLINK := $(GO_BASE_DIR)/go
GO_BIN := $(GO_SYMLINK)/bin/go
GO := $(GO_BIN)

export CGO_ENABLED := 0
export GOOS := $(OS)
export GOARCH := $(ARCH)
export PATH := $(BIN_DIR)/go/bin:$(BIN_DIR):$(PATH)

# ------------------------------------------------------------------------------
# BuildKit (Linux only)
# ------------------------------------------------------------------------------

BUILDKIT_VERSION := v0.27.1
BUILDKIT_TGZ := buildkit-$(BUILDKIT_VERSION).$(OS)-$(ARCH).tar.gz
BUILDKIT_URL := https://github.com/moby/buildkit/releases/download/$(BUILDKIT_VERSION)/$(BUILDKIT_TGZ)
BUILDCTL := $(BIN_DIR)/buildctl

# ------------------------------------------------------------------------------
# Kubernetes tooling
# ------------------------------------------------------------------------------

CONTROLLER_GEN := $(BIN_DIR)/controller-gen
KUSTOMIZE := $(BIN_DIR)/kustomize

# ------------------------------------------------------------------------------
# Docker detection
# ------------------------------------------------------------------------------

DOCKER ?= docker
DOCKER_BUILDX := $(shell docker buildx version >/dev/null 2>&1 && echo yes || echo no)
PLATFORMS ?= linux/amd64,linux/arm64

# ==============================================================================
# Toolchain installation
# ==============================================================================

.PHONY: go
go:
	@mkdir -p $(BIN_DIR)
	@if [ -x "$(GO_VERSION_DIR)/bin/go" ]; then \
		echo "✅ Go $(GO_VERSION) already installed"; \
	else \
		echo "⬇️ Installing Go $(GO_VERSION)"; \
		curl -fL $(GO_URL) -o /tmp/$(GO_TGZ); \
		tar -xzf /tmp/$(GO_TGZ) -C /tmp; \
		mv /tmp/go $(GO_VERSION_DIR); \
		rm -f /tmp/$(GO_TGZ); \
	fi
	@ln -sfn go-$(GO_VERSION) $(GO_SYMLINK)
	@$(GO_BIN) version

.PHONY: controller-gen
controller-gen: go
	@if [ ! -x "$(CONTROLLER_GEN)" ]; then \
		GOBIN=$(BIN_DIR) $(GO) install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0 ;\
	fi

.PHONY: deps
deps:
	$(GO) mod tidy
	$(GO) mod verify

.PHONY: go-cmd
go-cmd: go
	@if [ -z "$(ARGS)" ]; then \
		echo "Usage: make go-cmd ARGS=\"<go arguments>\""; \
		exit 1; \
	fi
	$(GO) $(ARGS)

.PHONY: envtest
envtest:
	@mkdir -p $(BIN_DIR)
	@if [ ! -x "$(ENVTEST)" ]; then \
		curl -fL $(ENVTEST_URL) -o $(ENVTEST); \
		chmod +x $(ENVTEST); \
	fi

# ==============================================================================
# Development
# ==============================================================================

.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object paths="./..."

.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd paths="./..." \
		output:crd:artifacts:config=manifests/charts/namespace-operator/crds

.PHONY: fmt
fmt: go
	$(GO) fmt ./...

.PHONY: lint
lint: vet fmt ## Same behavior as before

.PHONY: vet
vet: go
	$(GO) vet ./...

.PHONY: test
test: envtest generate manifests fmt vet
	$(ENVTEST) use latest --bin-dir $(BIN_DIR) >/dev/null
	KUBEBUILDER_ASSETS=$$($(ENVTEST) use $(K8S_VERSION) --bin-dir $(BIN_DIR) -p path) \
	$(GO) test ./... -coverprofile cover.out

# ==============================================================================
# Build
# ==============================================================================

.PHONY: build
build: fmt vet
	$(GO) build -o bin/manager main.go

.PHONY: run
run:
	$(GO) run ./main.go

# ==============================================================================
# Container image (Linux = BuildKit / macOS = Docker)
# ==============================================================================

.PHONY: image-build
image-build:
ifeq ($(OS),linux)
	@echo "🐧 Using BuildKit"
	@mkdir -p $(BIN_DIR)
	@if [ ! -x "$(BUILDCTL)" ]; then \
		curl -fL $(BUILDKIT_URL) -o /tmp/$(BUILDKIT_TGZ); \
		tar -xzf /tmp/$(BUILDKIT_TGZ) -C /tmp; \
		mv /tmp/bin/buildctl $(BUILDCTL); \
		chmod +x $(BUILDCTL); \
		rm -rf /tmp/$(BUILDKIT_TGZ) /tmp/bin; \
	fi
	sudo $(BUILDCTL) build \
	  --frontend dockerfile.v0 \
	  --local context=. \
	  --local dockerfile=. \
	  --output type=image,name=$(IMG),push=false
else
	@echo "🍎 Using Docker"
	$(DOCKER) build -t $(IMG) .
endif

.PHONY: image-push
image-push:
ifeq ($(OS),linux)
	@echo "🐧 Using BuildKit push"
	sudo $(BUILDCTL) build \
	  --frontend dockerfile.v0 \
	  --local context=. \
	  --local dockerfile=. \
	  --output type=image,name=$(IMG),push=true
else
	$(DOCKER) push $(IMG)
endif

# ==============================================================================
# Helm
# ==============================================================================

.PHONY: helm-install
helm-install:
	helm install $(HELM_RELEASE) $(HELM_CHART) \
	  --namespace $(HELM_NAMESPACE) --create-namespace

.PHONY: helm-upgrade
helm-upgrade:
	helm upgrade $(HELM_RELEASE) $(HELM_CHART) \
	  --namespace $(HELM_NAMESPACE)

.PHONY: helm-uninstall
helm-uninstall:
	helm uninstall $(HELM_RELEASE) \
	  --namespace $(HELM_NAMESPACE)

.PHONY: docs
docs:
	$(HELM_DOCS)

# ==============================================================================
# Info
# ==============================================================================

.PHONY: info
info:
	@echo "OS      : $(OS)"
	@echo "ARCH    : $(ARCH)"
	@echo "Go      : $(GO_VERSION)"
	@echo "Image   : $(IMG)"