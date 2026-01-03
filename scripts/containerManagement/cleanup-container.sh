#!/bin/bash
# scripts/containerManagement/cleanup-container.sh

set -euo pipefail

NAMESPACE="user-management"
IMAGE_NAME="user-management-service"
IMAGE_TAG="latest"
FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"

echo "🧹 Cleaning up User Management Service deployment..."

if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    echo "Deleting namespace '$NAMESPACE'..."
    kubectl delete namespace "$NAMESPACE"
    echo "✅ Namespace deleted."
else
    echo "⚠️ Namespace '$NAMESPACE' does not exist."
fi

echo "🗑️  Deleting Docker image '${FULL_IMAGE_NAME}' from Minikube environment..."
if command -v minikube >/dev/null 2>&1 && minikube status >/dev/null 2>&1; then
    eval "$(minikube docker-env)"
    if docker image inspect "$FULL_IMAGE_NAME" >/dev/null 2>&1; then
        docker rmi "$FULL_IMAGE_NAME"
        echo "✅ Docker image deleted."
    else
        echo "⚠️  Docker image '${FULL_IMAGE_NAME}' not found."
    fi
else
    echo "⚠️  Minikube is not running or not installed. Skipping image deletion."
fi

echo "✅ Cleanup complete."
