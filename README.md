# Namespace Operator

A Kubernetes operator that manages **namespaces, resource quotas, and limit ranges**
using a custom resource called **Tenant**.

The operator reconciles cluster-scoped `Tenant` resources and ensures that the
corresponding Kubernetes namespaces are created and configured according to
the declared specifications.

---

## ✨ Features

- Automatic namespace creation
- ResourceQuota management per tenant
- LimitRange management per tenant
- Cluster-scoped custom resource (`Tenant`)
- Continuous reconciliation (self-healing)

---

## 📦 Architecture Overview

- **Custom Resource**: `Tenant`
- **API Group**: `platform.example.com`
- **Version**: `v1alpha1`
- **Scope**: Cluster
- **Controller framework**: controller-runtime

---

## 📁 Repository Structure

```
namespace-operator/
├── api/                    # API types (Tenant CRD)
├── controllers/            # Reconciler logic
├── manifests/              # Kubernetes manifests (CRDs, packaging, etc.)
├── Dockerfile              # Operator container image
├── Makefile
└── README.md
```

---

## 📘 Custom Resource: Tenant

The `Tenant` resource represents a logical tenant in the cluster.

### Example

```yaml
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
    defaultCpu: "200m"
    defaultMemory: "256Mi"
    maxCpu: "1"
    maxMemory: "1Gi"
```

Apply it:

```bash
kubectl apply -f tenant.yaml
```

---

## 🔍 Reconciliation Behavior

When a `Tenant` resource is created or updated, the operator:

1. Ensures the target namespace exists
2. Creates or updates a `ResourceQuota`
3. Creates or updates a `LimitRange`
4. Continuously reconciles to enforce the desired state

If resources are modified or deleted manually, the operator will restore them.

---

## 🛠 Development

### Run tests

```bash
make test
```

---

### Build the operator image

```bash
make image-build
make image-push
```

---

### Run locally (against a cluster)

```bash
make run
```

> The operator will use your current kubeconfig.

---

## 📌 Notes

- The operator is **cluster-scoped**
- It requires permissions to manage:
  - Namespaces
  - ResourceQuotas
  - LimitRanges
  - Tenant custom resources
- Deployment and packaging are intentionally not documented here

---

## 📄 License

Apache 2.0
