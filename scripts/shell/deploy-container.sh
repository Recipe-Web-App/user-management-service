#!/bin/bash
# scripts/shell/deploy-container.sh

set -euo pipefail

NAMESPACE="user-management"
CONFIG_DIR="k8s"
SECRET_NAME="user-management-db-password"
IMAGE_NAME="user-management-service"
IMAGE_TAG="latest"
FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"

# Fixes bug where first separator line does not fill the terminal width
COLUMNS=$(tput cols 2>/dev/null || echo 80)

# Utility function for printing section separators
print_separator() {
  local char="${1:-=}"
  local width="${COLUMNS:-80}"
  printf '%*s\n' "$width" '' | tr ' ' "$char"
}

print_separator "="
echo "🔧 Setting up Minikube environment..."
print_separator "-"
env_status=true
if ! command -v minikube >/dev/null 2>&1; then
  echo "❌ Minikube is not installed. Please install it first."
  env_status=false
else
  echo "✅ Minikube is installed."
fi

if ! command -v kubectl >/dev/null 2>&1; then
  echo "❌ kubectl is not installed. Please install it first."
  env_status=false
else
  echo "✅ kubectl is installed."
fi
if ! command -v docker >/dev/null 2>&1; then
  echo "❌ Docker is not installed. Please install it first."
  env_status=false
else
  echo "✅ Docker is installed."
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "❌ jq is not installed. Please install it first."
  env_status=false
else
  echo "✅ jq is installed."
fi
if ! $env_status; then
  echo "Please resolve the above issues before proceeding."
  exit 1
fi

if ! minikube status >/dev/null 2>&1; then
  print_separator "-"
  echo "🚀 Starting Minikube..."
  minikube start

  echo "✅ Minikube started."
else
  echo "✅ Minikube is already running."
fi

print_separator "="
echo "📂 Ensuring namespace '${NAMESPACE}' exists..."
print_separator "-"

if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    echo "✅ '$NAMESPACE' namespace already exists."
else
    kubectl create namespace "$NAMESPACE"
    echo "✅ '$NAMESPACE' namespace created."
fi


print_separator "="
echo "🔧 Loading environment variables from .env file (if present)..."
print_separator "-"

if [ -f .env ]; then
    set -o allexport
    # Capture env before
    BEFORE_ENV=$(mktemp)
    AFTER_ENV=$(mktemp)
    env | cut -d= -f1 | sort > "$BEFORE_ENV"
    # shellcheck source=.env disable=SC1091
    source .env
    # Capture env after
    env | cut -d= -f1 | sort > "$AFTER_ENV"
    # Show newly loaded/changed variables
    echo "✅ Loaded variables from .env:"
    comm -13 "$BEFORE_ENV" "$AFTER_ENV"
    rm -f "$BEFORE_ENV" "$AFTER_ENV"
    set +o allexport
fi

print_separator "="
echo "🐳 Building Docker image: ${FULL_IMAGE_NAME} (inside Minikube Docker daemon)"
print_separator '-'

eval "$(minikube docker-env)"
docker build -t "$FULL_IMAGE_NAME" .
echo "✅ Docker image '${FULL_IMAGE_NAME}' built successfully."

print_separator "="
echo "⚙️ Creating/Updating ConfigMap from env..."
print_separator "-"

envsubst < "${CONFIG_DIR}/configmap-template.yaml" | kubectl apply -f -

print_separator "="
echo "🔐 Creating/updating Secret..."
print_separator "-"

kubectl delete secret "$SECRET_NAME" -n "$NAMESPACE" --ignore-not-found
envsubst < "${CONFIG_DIR}/secret-template.yaml" | kubectl apply -f -

print_separator "="
echo "📦 Deploying User Management Service container..."
print_separator "-"

kubectl apply -f "${CONFIG_DIR}/deployment.yaml"

print_separator "="
echo "🌐 Exposing User Management Service via ClusterIP Service..."
print_separator "-"

kubectl apply -f "${CONFIG_DIR}/service.yaml"

print_separator "="
echo "📥 Applying HTTPRoute for Kong Gateway..."
print_separator "-"

kubectl apply -f "${CONFIG_DIR}/gateway-route.yaml"

print_separator "="
echo "⏳ Waiting for User Management Service pod to be ready..."
print_separator "-"

kubectl wait --namespace="$NAMESPACE" \
  --for=condition=Ready pod \
  --selector=app=user-management \
  --timeout=90s

print_separator "-"
echo "✅ User Management Service is up and running in namespace '$NAMESPACE'."

print_separator "="
echo "🔗 Setting up /etc/hosts for sous-chef-proxy.local..."
print_separator "-"

MINIKUBE_IP=$(minikube ip)
if grep -q "user-management.local" /etc/hosts; then
  echo "🔄 Updating /etc/hosts for user-management.local..."
  sed -i "/user-management.local/d" /etc/hosts
else
  echo "➕ Adding user-management.local to /etc/hosts..."
fi
echo "$MINIKUBE_IP user-management.local" >> /etc/hosts
echo "✅ /etc/hosts updated with user-management.local pointing to $MINIKUBE_IP"

print_separator "="
echo "🌍 You can now access your app at: http://sous-chef-proxy.local/api/v1/user-management/health"

POD_NAME=$(kubectl get pods -n "$NAMESPACE" -l app=user-management -o jsonpath="{.items[0].metadata.name}")
SERVICE_JSON=$(kubectl get svc user-management -n "$NAMESPACE" -o json)
SERVICE_IP=$(echo "$SERVICE_JSON" | jq -r '.spec.clusterIP')
SERVICE_PORT=$(echo "$SERVICE_JSON" | jq -r '.spec.ports[0].port')
HTTPROUTE_HOSTS=$(kubectl get httproute -n "$NAMESPACE" -o jsonpath='{.items[*].spec.hostnames[*]}' | tr ' ' '\n' | sort -u | paste -sd ',' -)

print_separator "="
echo "📡 Access info:"
echo "  Pod: $POD_NAME"
echo "  Service: $SERVICE_IP:$SERVICE_PORT"
echo "  HTTPRoute Hosts: $HTTPROUTE_HOSTS"
echo "  Gateway: kong (namespace: kong)"
echo "  Health Check: http://sous-chef-proxy.local/api/v1/user-management/health"
print_separator "="
