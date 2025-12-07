# User Management Service

A Golang-based service for managing user data.

## Deployment

### Prerequisites

- Docker
- Minikube
- Kubectl

### Local Deployment (Minikube)

We provide shell scripts to manage the lifecycle of the application in Minikube.

1. **Start/Deploy**:

    ```bash
    ./scripts/containerManagement/deploy-container.sh
    ```

    This script will:
    - Check for required tools.
    - Start Minikube if not running.
    - Build the Docker image inside Minikube's Docker daemon.
    - Apply ConfigMaps, Secrets, and Deployments.
    - Wait for the pod to be ready.

2. **Monitor**:

    ```bash
    ./scripts/containerManagement/get-container-status.sh
    ```

    Displays a dashboard with pod status, logs, resource usage, and health checks.

3. **Update**:

    ```bash
    ./scripts/containerManagement/update-container.sh
    ```

    updates the image and performs a rollout restart.

4. **Stop/Start**:

    ```bash
    ./scripts/containerManagement/stop-container.sh
    ./scripts/containerManagement/start-container.sh
    ```

    Easily pause and resume the service without full redeployment.

5. **Cleanup**:

    ```bash
    ./scripts/containerManagement/cleanup-container.sh
    ```

    Removes the namespace and all associated resources.

### Kubernetes Resources

Manifests are located in the `k8s/` directory:

- `namespace.yaml`: Defines `user-management-dev-poc`.
- `deployment.yaml`: Main application deployment with probes.
- `service.yaml`: ClusterIP service.
- `configmap-template.yaml`: Configuration template.
- `secret-template.yaml`: Secret template.

### Configuration

Configuration is managed via `config/` files and Kubernetes ConfigMaps/Secrets.
The application loads config from `config/` and overrides with environment variables.
