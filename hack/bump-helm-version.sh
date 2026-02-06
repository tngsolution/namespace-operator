#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-}"

if [[ -z "$TAG" ]]; then
  echo "Usage: $0 vX.Y.Z"
  exit 1
fi

VERSION="${TAG#v}"

CHART="manifests/charts/namespace-operator/Chart.yaml"
VALUES="manifests/charts/namespace-operator/values.yaml"

echo "🔖 Bumping version to ${VERSION}"

# Chart.yaml
yq -i "
  .version = \"${VERSION}\" |
  .appVersion = \"${VERSION}\"
" "$CHART"

# values.yaml (image tag uniquement)
yq -i "
  .image.tag = \"${VERSION}\"
" "$VALUES"

echo "✅ Chart.yaml & values.yaml updated"