#!/bin/bash
# scripts/containerManagement/deploy-container.sh

set -euo pipefail

NAMESPACE="user-management"
CONFIG_DIR="k8s"
SECRET_NAME="user-management-secrets" # pragma: allowlist secret
IMAGE_NAME="user-management-service"
IMAGE_TAG="latest"
FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"

COLUMNS=$(tput cols 2>/dev/null || echo 80)

# Colors for better readability
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_separator() {
  local char="${1:-=}"
  local width="${COLUMNS:-80}"
  printf '%*s\n' "$width" '' | tr ' ' "$char"
}

# Function to print status with color
print_status() {
    local status="$1"
    local message="$2"
    if [ "$status" = "ok" ]; then
        echo -e "✅ ${GREEN}$message${NC}"
    elif [ "$status" = "warning" ]; then
        echo -e "⚠️  ${YELLOW}$message${NC}"
    else
        echo -e "❌ ${RED}$message${NC}"
    fi
}

print_separator "="
echo -e "${CYAN}🔧 Setting up Minikube environment...${NC}"
print_separator "-"
env_status=true
if ! command -v minikube >/dev/null 2>&1; then
  print_status "error" "Minikube is not installed. Please install it first."
  env_status=false
else
  print_status "ok" "Minikube is installed."
fi

if ! command -v kubectl >/dev/null 2>&1; then
  print_status "error" "kubectl is not installed. Please install it first."
  env_status=false
else
  print_status "ok" "kubectl is installed."
fi
if ! command -v docker >/dev/null 2>&1; then
  print_status "error" "Docker is not installed. Please install it first."
  env_status=false
else
  print_status "ok" "Docker is installed."
fi
if ! command -v jq >/dev/null 2>&1; then
  print_status "error" "jq is not installed. Please install it first."
  env_status=false
else
  print_status "ok" "jq is installed."
fi
if ! $env_status; then
  echo "Please resolve the above issues before proceeding."
  exit 1
fi

if ! minikube status >/dev/null 2>&1; then
  print_separator "-"
  echo -e "${YELLOW}🚀 Starting Minikube...${NC}"
  minikube start
else
  print_status "ok" "Minikube is already running."
fi

print_separator "="
echo -e "${CYAN}📂 Ensuring namespace '${NAMESPACE}' exists...${NC}"
print_separator "-"

if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    print_status "ok" "'$NAMESPACE' namespace already exists."
else
    # We apply the namespace manifest first to ensure it's created as defined
    kubectl apply -f "${CONFIG_DIR}/namespace.yaml"
    print_status "ok" "'$NAMESPACE' namespace created."
fi

print_separator "="
echo -e "${CYAN}🔧 Loading environment variables from .env.prod file (if present)...${NC}"
print_separator "-"

if [ -f .env.prod ]; then
    set -o allexport
    BEFORE_ENV=$(mktemp)
    AFTER_ENV=$(mktemp)
    env | cut -d= -f1 | sort > "$BEFORE_ENV"
    # shellcheck source=.env.prod disable=SC1091
    source .env.prod
    env | cut -d= -f1 | sort > "$AFTER_ENV"
    print_status "ok" "Loaded variables from .env.prod:"
    comm -13 "$BEFORE_ENV" "$AFTER_ENV"
    rm -f "$BEFORE_ENV" "$AFTER_ENV"
    set +o allexport
else
    print_status "warning" ".env.prod file not found, using existing environment variables"
fi

print_separator "="
echo -e "${CYAN}🟢 Building Go Service Docker image: ${FULL_IMAGE_NAME} (inside Minikube Docker daemon)${NC}"
print_separator '-'

eval "$(minikube docker-env)"
docker build -t "$FULL_IMAGE_NAME" .
print_status "ok" "Docker image '${FULL_IMAGE_NAME}' built successfully."

print_separator "="
echo -e "${CYAN}⚙️  Creating/Updating ConfigMap from env...${NC}"
print_separator "-"

envsubst < "${CONFIG_DIR}/configmap-template.yaml" | kubectl apply -f -

print_separator "="
echo -e "${CYAN}🔐 Creating/updating Secret...${NC}"
print_separator "-"

kubectl delete secret "$SECRET_NAME" -n "$NAMESPACE" --ignore-not-found
envsubst < "${CONFIG_DIR}/secret-template.yaml" | kubectl apply -f -

print_separator "="
echo -e "${CYAN}📦 Deploying User Management Service...${NC}"
print_separator "-"

kubectl apply -f "${CONFIG_DIR}/deployment.yaml"

print_separator "="
echo -e "${CYAN}🌐 Exposing User Management Service via ClusterIP Service...${NC}"
print_separator "-"

kubectl apply -f "${CONFIG_DIR}/service.yaml"

print_separator "="
echo -e "${CYAN}⏳ Waiting for User Management Service pod to be ready...${NC}"
print_separator "-"

kubectl wait --namespace="$NAMESPACE" \
  --for=condition=Ready pod \
  --selector=app=user-management-service \
  --timeout=90s

print_separator "-"
print_status "ok" "User Management Service is up and running in namespace '$NAMESPACE'."

print_separator "="
echo -e "${CYAN}🔗 Setting up /etc/hosts for user-management.local...${NC}"
print_separator "-"

MINIKUBE_IP=$(minikube ip)
if grep -q "user-management.local" /etc/hosts; then
  echo -e "${YELLOW}🔄 Updating /etc/hosts for user-management.local...${NC}"
  sed -i "/user-management.local/d" /etc/hosts
else
  echo -e "${GREEN}➕ Adding user-management.local to /etc/hosts...${NC}"
fi
echo "$MINIKUBE_IP user-management.local" | tee -a /etc/hosts > /dev/null
print_status "ok" "/etc/hosts updated with user-management.local pointing to $MINIKUBE_IP"

print_separator "="
echo -e "${CYAN}🛰️  Access info:${NC}"

POD_NAME=$(kubectl get pods -n "$NAMESPACE" -l app=user-management-service -o jsonpath="{.items[0].metadata.name}")
SERVICE_JSON=$(kubectl get svc user-management-service -n "$NAMESPACE" -o json)
SERVICE_IP=$(echo "$SERVICE_JSON" | jq -r '.spec.clusterIP')
SERVICE_PORT=$(echo "$SERVICE_JSON" | jq -r '.spec.ports[0].port')

echo "  Pod: $POD_NAME"
echo "  Service: $SERVICE_IP:$SERVICE_PORT"
echo "  Health Check: http://user-management.local/api/v1/user-management/health"
echo "  To access locally without gateway: kubectl port-forward -n $NAMESPACE svc/user-management-service 8080:8080"
print_separator "="
