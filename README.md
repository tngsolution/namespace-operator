# Namespace Operator

A Kubernetes operator that manages **namespaces, resource quotas, limit
ranges, and network policies** using a custom resource called
**Tenant**.

The operator reconciles cluster-scoped `Tenant` resources and ensures
that the corresponding Kubernetes namespaces are created and configured
according to the declared specifications.

------------------------------------------------------------------------

## ✨ Features

-   Automatic namespace creation\
-   ResourceQuota management per tenant\
-   LimitRange management per tenant\
-   NetworkPolicy management per tenant\
-   Optional reusable configuration via `TenantProfile`\
-   Cluster-scoped custom resource (`Tenant`)\
-   Continuous reconciliation (self-healing)\
-   GitOps-ready structure (ArgoCD compatible)

------------------------------------------------------------------------

## 📦 Architecture Overview

-   **Custom Resource**: `Tenant`
-   **Optional Template Resource**: `TenantProfile`
-   **API Group**: `platform.example.com`
-   **Version**: `v1alpha1`
-   **Scope**: Cluster
-   **Controller framework**: controller-runtime

------------------------------------------------------------------------

## 📁 Repository Structure

    namespace-operator/
    ├── src/                         # Operator source code (Go module)
    │   ├── api/                     # CRD types (Tenant, TenantProfile)
    │   ├── controllers/             # Reconciler logic
    │   ├── main.go
    │   ├── go.mod
    │   └── go.sum
    │
    ├── manifests/                   # Helm chart and CRDs
    │   └── charts/namespace-operator
    │
    ├── argocd/                      # ArgoCD Applications
    │
    ├── gitops/                      # Declarative platform configuration
    │   ├── profiles/                # TenantProfile resources
    │   └── tenants/                 # Tenant resources
    │
    ├── Makefile
    └── README.md

------------------------------------------------------------------------

## 📘 Custom Resource: Tenant

The `Tenant` resource represents a logical tenant in the cluster.

It defines:

-   Target namespace
-   Resource quotas
-   Limit ranges
-   Optional network policies
-   Optional profile reference

------------------------------------------------------------------------

### Example (inline configuration)

``` yaml
apiVersion: platform.example.com/v1alpha1
kind: Tenant
metadata:
  name: team-devops
spec:
  namespace: team-devops
  quota:
    cpu: "2"
    memory: "4Gi"
    pods: 10
  limits:
    defaultCPU: "200m"
    defaultMemory: "256Mi"
    maxCPU: "1"
    maxMemory: "1Gi"
```

Apply it:

``` bash
kubectl apply -f tenant.yaml
```

------------------------------------------------------------------------

## 📘 Custom Resource: TenantProfile

Reusable configuration template for quotas and limits.

``` yaml
apiVersion: platform.example.com/v1alpha1
kind: TenantProfile
metadata:
  name: small
spec:
  quota:
    cpu: "1"
    memory: "2Gi"
    pods: 5
  limits:
    defaultCPU: "100m"
    defaultMemory: "128Mi"
    maxCPU: "500m"
    maxMemory: "512Mi"
```

Then reference it from a Tenant:

``` yaml
spec:
  namespace: team-a
  profile: small
```

The profile takes precedence over inline configuration.

------------------------------------------------------------------------

## 🔍 Reconciliation Behavior

When a `Tenant` resource is created or updated, the operator:

1.  Ensures the target namespace exists\
2.  Applies ResourceQuota\
3.  Applies LimitRange\
4.  Applies NetworkPolicies (default-deny or custom rules)\
5.  Updates the Tenant status

If resources are modified or deleted manually, the operator restores
them to the desired state.

On Tenant deletion:

-   The namespace is deleted
-   Finalizers ensure clean teardown

------------------------------------------------------------------------

## 🔁 GitOps Integration

This repository is designed to work with ArgoCD:

-   `argocd/` contains ArgoCD Application manifests
-   `gitops/profiles/` contains TenantProfiles
-   `gitops/tenants/` contains Tenant resources

Typical flow:

Git → ArgoCD → Kubernetes Cluster

Any change committed to Git is automatically reconciled in the cluster.

------------------------------------------------------------------------

## 🛠 Development

### Run tests

``` bash
make test
```

------------------------------------------------------------------------

### Build the operator image

``` bash
make image-build
make image-push
```

------------------------------------------------------------------------

### Run locally (against a cluster)

``` bash
make run
```

The operator uses your current kubeconfig.

------------------------------------------------------------------------

## 📌 Notes

-   The operator is **cluster-scoped**
-   It requires permissions to manage:
    -   Namespaces
    -   ResourceQuotas
    -   LimitRanges
    -   NetworkPolicies
    -   Tenant and TenantProfile resources

------------------------------------------------------------------------

## 📄 License

Apache 2.0
