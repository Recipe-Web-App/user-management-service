#!/bin/bash
# scripts/containerManagement/get-container-status.sh

set -euo pipefail

NAMESPACE="user-management-dev-poc"

# Colors for better readability
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
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

# Function to print separator
print_separator() {
    local char="${1:-─}"
    local width="${2:-$(tput cols 2>/dev/null || echo 80)}"
    printf "%*s\n" "$width" '' | tr ' ' "$char"
}

# Function to test HTTP endpoint with timing
test_endpoint() {
    local url="$1"
    local description="$2"
    local timeout="${3:-5}"

    echo -e "${BLUE}  Testing: $description${NC}"

    if command_exists curl; then
        local response
        response=$(curl -s -w "%{http_code},%{time_total}" -m "$timeout" "$url" 2>/dev/null || echo "000,0.000")

        local http_code
        http_code=$(echo "$response" | tail -1 | cut -d',' -f1)
        local response_time
        response_time=$(echo "$response" | tail -1 | cut -d',' -f2)
        local body
        body=$(echo "$response" | head -n -1)

        if [ "$http_code" = "200" ]; then
            echo -e "    ✅ ${GREEN}HTTP $http_code${NC} - Response time: ${response_time}s"
            if echo "$body" | jq . >/dev/null 2>&1; then
                local status
                status=$(echo "$body" | jq -r '.status // empty')
                if [ -n "$status" ]; then
                    echo -e "    📊 Status: ${GREEN}$status${NC}"
                fi
            fi
        elif [ "$http_code" = "503" ]; then
            echo -e "    ⚠️  ${YELLOW}HTTP $http_code${NC} - Service unavailable - Response time: ${response_time}s"
        elif [ "$http_code" = "000" ]; then
            echo -e "    ❌ ${RED}Connection failed${NC} - Timeout or unreachable"
        else
            echo -e "    ❌ ${RED}HTTP $http_code${NC} - Response time: ${response_time}s"
        fi
    else
        echo -e "    ⚠️  ${YELLOW}curl not available - cannot test endpoint${NC}"
    fi
}

echo "📊 User Management Service Status Dashboard"
print_separator "="

# Check prerequisites
echo ""
echo -e "${CYAN}🔧 Prerequisites Check:${NC}"
for cmd in kubectl minikube curl jq; do
    if command_exists "$cmd"; then
        print_status "ok" "$cmd is available"
    else
        print_status "warning" "$cmd is not installed"
    fi
done

if command_exists minikube; then
    if minikube status >/dev/null 2>&1; then
        print_status "ok" "minikube is running"
    else
        print_status "warning" "minikube is not running"
    fi
fi

print_separator
echo ""
echo -e "${CYAN}🔍 Namespace Status:${NC}"
if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    print_status "ok" "Namespace '$NAMESPACE' exists"
    NAMESPACE_AGE=$(kubectl get namespace "$NAMESPACE" -o jsonpath='{.metadata.creationTimestamp}' | xargs -I {} date -d {} "+%Y-%m-%d %H:%M:%S" 2>/dev/null || echo "unknown")
    RESOURCE_COUNT=$(kubectl get all -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l || echo "unknown")
    echo "   📅 Created: $NAMESPACE_AGE, Resources: $RESOURCE_COUNT"
else
    print_status "error" "Namespace '$NAMESPACE' does not exist"
    echo -e "${YELLOW}💡 Run ./scripts/containerManagement/deploy-container.sh to deploy${NC}"
    exit 1
fi

print_separator
echo ""
echo -e "${CYAN}📦 Deployment Status:${NC}"
if kubectl get deployment user-management-service -n "$NAMESPACE" >/dev/null 2>&1; then
    kubectl get deployment user-management-service -n "$NAMESPACE"

    READY_REPLICAS=$(kubectl get deployment user-management-service -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
    DESIRED_REPLICAS=$(kubectl get deployment user-management-service -n "$NAMESPACE" -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")

    if [ "$READY_REPLICAS" = "$DESIRED_REPLICAS" ] && [ "$READY_REPLICAS" != "0" ]; then
        print_status "ok" "Deployment is ready ($READY_REPLICAS/$DESIRED_REPLICAS replicas)"
    else
        print_status "warning" "Deployment not fully ready ($READY_REPLICAS/$DESIRED_REPLICAS replicas)"
    fi
else
    print_status "error" "Deployment not found"
fi

print_separator
echo ""
echo -e "${CYAN}🐳 Pod Status:${NC}"
if kubectl get pods -n "$NAMESPACE" -l app=user-management-service >/dev/null 2>&1; then
    kubectl get pods -n "$NAMESPACE" -l app=user-management-service

    POD_NAME=$(kubectl get pods -n "$NAMESPACE" -l app=user-management-service -o jsonpath="{.items[0].metadata.name}" 2>/dev/null || echo "")
    if [ -n "$POD_NAME" ]; then
        echo ""
        echo -e "${CYAN}📋 Pod Details:${NC}"
        kubectl describe pod "$POD_NAME" -n "$NAMESPACE" | grep -A5 -E "Conditions:|Events:" || true
    fi
else
    print_status "error" "No pods found"
fi

print_separator
echo ""
echo -e "${CYAN}🏥 Health Check Dashboard:${NC}"
if [ -n "$POD_NAME" ]; then
    # Since we don't have Ingress yet, we'll check via port-forward logic conceptually or via kubectl exec
    # For now, let's use kubectl exec to check localhost inside the pod if curl is available there,
    # OR warn that external access requires port-forwarding.
    # The reference script assumes /etc/hosts or ingress. We only have ClusterIP.

    echo -e "${PURPLE}🔍 Checking connectivity...${NC}"

    # We can try to hit the pod IP from Minikube if available, or just describe the probes status
    # Let's rely on K8s probe status reported in Pod Details above, and provide manual instructions.

    echo "   To test endpoints locally:"
    echo "   kubectl port-forward -n $NAMESPACE svc/user-management-service 8080:8080"
    echo "   Then curl http://localhost:8080/api/v1/user-management/health"
else
    print_status "warning" "Cannot check health - Pod not running"
fi

print_separator
echo ""
echo -e "${CYAN}🌐 Service Status:${NC}"
if kubectl get service user-management-service -n "$NAMESPACE" >/dev/null 2>&1; then
    kubectl get service user-management-service -n "$NAMESPACE"
    print_status "ok" "Service is available"
else
    print_status "error" "Service not found"
fi

print_separator
echo ""
echo -e "${CYAN}🔐 ConfigMap and Secret Status:${NC}"
if kubectl get configmap user-management-config -n "$NAMESPACE" >/dev/null 2>&1; then
    print_status "ok" "ConfigMap exists"
    CONFIG_KEYS=$(kubectl get configmap user-management-config -n "$NAMESPACE" -o jsonpath='{.data}' | jq -r 'keys[]' 2>/dev/null | wc -l || echo "0")
    echo "   🔑 Configuration keys: $CONFIG_KEYS"
else
    print_status "error" "ConfigMap not found"
fi

if kubectl get secret user-management-secrets -n "$NAMESPACE" >/dev/null 2>&1; then
    print_status "ok" "Secret exists"
    SECRET_KEYS=$(kubectl get secret user-management-secrets -n "$NAMESPACE" -o jsonpath='{.data}' | jq -r 'keys[]' 2>/dev/null | wc -l || echo "0")
    echo "   🔐 Secret keys: $SECRET_KEYS"
else
    print_status "error" "Secret not found"
fi

print_separator
echo ""
echo -e "${CYAN}🛡️  Security Posture Check:${NC}"
if [ -n "$POD_NAME" ]; then
    echo -e "${BLUE}  Checking pod security context...${NC}"

    RUN_AS_NON_ROOT=$(kubectl get pod "$POD_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.securityContext.runAsNonRoot}' 2>/dev/null || echo "false")
    # For container[0]
    READ_ONLY_ROOT=$(kubectl get pod "$POD_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.containers[0].securityContext.readOnlyRootFilesystem}' 2>/dev/null || echo "false")

    if [ "$RUN_AS_NON_ROOT" = "true" ]; then
        print_status "ok" "Running as non-root user"
    else
        print_status "warning" "Not explicitly set to run as non-root"
    fi

    if [ "$READ_ONLY_ROOT" = "true" ]; then
        print_status "ok" "Read-only root filesystem enabled"
    else
        print_status "warning" "Read-only root filesystem not enabled"
    fi
fi

print_separator
echo ""
echo -e "${CYAN}📜 Recent Events & Troubleshooting:${NC}"
echo -e "${BLUE}  Recent pod events:${NC}"
kubectl get events -n "$NAMESPACE" --sort-by='.lastTimestamp' --field-selector involvedObject.kind=Pod | tail -5 || print_status "warning" "No recent events found"

if [ -n "$POD_NAME" ]; then
    echo ""
    echo -e "${BLUE}  Container logs (last 10 lines):${NC}"
    kubectl logs "$POD_NAME" -n "$NAMESPACE" --tail=10 2>/dev/null || print_status "warning" "Logs not available"

    echo ""
    echo -e "${BLUE}  Container restart count:${NC}"
    RESTART_COUNT=$(kubectl get pod "$POD_NAME" -n "$NAMESPACE" -o jsonpath='{.status.containerStatuses[0].restartCount}' 2>/dev/null || echo "unknown")
    if [ "$RESTART_COUNT" = "0" ]; then
        print_status "ok" "No restarts: $RESTART_COUNT"
    elif [ "$RESTART_COUNT" != "unknown" ] && [ "$RESTART_COUNT" -lt 3 ]; then
        print_status "warning" "Low restart count: $RESTART_COUNT"
    else
        print_status "error" "High restart count: $RESTART_COUNT"
    fi
fi

print_separator "="
echo -e "${GREEN}📊 Status check completed!${NC}"
echo -e "${CYAN}💡 Quick actions:${NC}"
echo "   🚀 Start: ./scripts/containerManagement/start-container.sh"
echo "   🛑 Stop: ./scripts/containerManagement/stop-container.sh"
echo "   🔄 Update: ./scripts/containerManagement/update-container.sh"
echo "   🧹 Cleanup: ./scripts/containerManagement/cleanup-container.sh"
print_separator "="
