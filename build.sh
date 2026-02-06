#!/usr/bin/env bash
#set -euo pipefail
export IMG=docker.io/baabdoul/namespace-operator:1.0.2

# =========================
# Preconditions
# =========================
if [[ -z "${IMG:-}" ]]; then
  echo "❌ IMG is not set"
  echo "👉 export IMG=docker.io/baabdoul/namespace-operator:1.0.2"
  exit 1
fi

echo "🚀 Building image via Makefile"
echo "📦 Image: $IMG"

# =========================
# Build via Makefile
# =========================
make image-build
make image-push

echo "✅ Build & push completed: $IMG"
