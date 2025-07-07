# Container Management Scripts

This directory contains Python scripts for managing the user-management-service deployment in Kubernetes using Minikube.

## Prerequisites

Before using these scripts, ensure you have the following installed:

- **Minikube**: Local Kubernetes cluster
- **kubectl**: Kubernetes command-line tool
- **Docker**: Container runtime
- **Python 3.8+**: Python runtime
- **Poetry**: Python dependency management

## Installation

Install the required dependencies using Poetry:

```bash
poetry install
```

## Scripts Overview

### 1. Deploy Container (`deploy_container.py`)

Deploys the user-management-service to Kubernetes.

**Usage:**

```bash
# Basic deployment
python scripts/containerManagement/deploy_container.py

# Custom namespace and image tag
python scripts/containerManagement/deploy_container.py --namespace my-namespace --image-tag v1.0.0

# Dry run (show what would be deployed)
python scripts/containerManagement/deploy_container.py --dry-run
```

**Features:**

- Checks prerequisites (Minikube, kubectl, Docker)
- Starts Minikube if not running
- Loads environment variables from `.env` file
- Builds Docker image
- Creates/updates Kubernetes resources
- Waits for deployment to be ready
- Provides service URL

### 2. Get Container Status (`get_container_status.py`)

Checks the status of the user-management-service deployment.

**Usage:**

```bash
# Basic status check
python scripts/containerManagement/get_container_status.py

# Custom namespace
python scripts/containerManagement/get_container_status.py --namespace my-namespace

# Detailed information
python scripts/containerManagement/get_container_status.py --detailed
```

**Features:**

- Minikube status
- Namespace status
- Deployment status
- Service status
- Ingress status
- Pod status
- PVC status

### 3. Start Container (`start_container.py`)

Starts the user-management-service deployment by scaling to 1 replica.

**Usage:**

```bash
# Basic start
python scripts/containerManagement/start_container.py

# Custom namespace and deployment
python scripts/containerManagement/start_container.py --namespace my-namespace --deployment my-deployment

# Custom replicas and timeout
python scripts/containerManagement/start_container.py --replicas 3 --timeout 120
```

**Features:**

- Checks Minikube status
- Verifies deployment exists
- Scales deployment to specified replicas
- Waits for deployment to be ready
- Shows final deployment status

### 4. Stop Container (`stop_container.py`)

Stops the user-management-service deployment by scaling to 0 replicas.

**Usage:**

```bash
# Basic stop (scale to 0)
python scripts/containerManagement/stop_container.py

# Custom namespace and deployment
python scripts/containerManagement/stop_container.py --namespace my-namespace --deployment my-deployment

# Scale to specific number of replicas
python scripts/containerManagement/stop_container.py --replicas 2
```

**Features:**

- Checks Minikube status
- Verifies deployment exists
- Scales deployment to specified replicas
- Shows final deployment status

### 5. Cleanup Container (`cleanup_container.py`)

Completely removes the user-management-service deployment and related resources.

**Usage:**

```bash
# Interactive cleanup
python scripts/containerManagement/cleanup_container.py

# Force cleanup (skip confirmations)
python scripts/containerManagement/cleanup_container.py --force

# Custom namespace and config directory
python scripts/containerManagement/cleanup_container.py --namespace my-namespace --config-dir my-k8s

# Stop Minikube after cleanup
python scripts/containerManagement/cleanup_container.py --stop-minikube
```

**Features:**

- Starts Minikube if not running
- Deletes Kubernetes resources (deployment, service, ingress, etc.)
- Deletes Kubernetes jobs
- Optionally deletes PVCs (with confirmation)
- Removes Docker image
- Optionally stops Minikube

## Environment Variables

The scripts use environment variables from a `.env` file. Create a `.env` file in the project root with the following variables:

```env
# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=user_management
POSTGRES_SCHEMA=public

# User Management Database
USER_MANAGEMENT_DB_USER=user_management_user
USER_MANAGEMENT_DB_PASSWORD=your_secure_password

# JWT Configuration
JWT_SECRET_KEY=your_jwt_secret_key
JWT_SIGNING_ALGORITHM=HS256
ACCESS_TOKEN_EXPIRE_MINUTES=30

# CORS Configuration
ALLOWED_ORIGIN_HOSTS=http://localhost:3000,http://localhost:8080
ALLOWED_CREDENTIALS=true
```

## Kubernetes Manifests

The scripts expect Kubernetes manifests in the `k8s/` directory:

- `configmap-template.yaml`: Application configuration
- `secret-template.yaml`: Database credentials
- `deployment.yaml`: Application deployment
- `service.yaml`: Service definition
- `ingress.yaml`: Ingress configuration
- `pvc.yaml`: Persistent volume claim

## Error Handling

All scripts include comprehensive error handling:

- **Prerequisites Check**: Verifies required tools are installed
- **Kubernetes Client**: Handles connection issues gracefully
- **Resource Operations**: Continues on non-critical failures
- **User Confirmation**: Prompts for destructive operations
- **Progress Tracking**: Shows real-time progress with spinners

## Rich Terminal Interface

The scripts use the Rich library for beautiful terminal output:

- **Color-coded Status**: Green for success, red for errors, yellow for warnings
- **Progress Bars**: Real-time progress tracking
- **Tables**: Structured data display
- **Panels**: Organized information sections
- **Spinners**: Animated progress indicators

## Troubleshooting

### Common Issues

1. **Minikube not running**

   ```bash
   minikube start
   ```

2. **Kubernetes client connection issues**

   ```bash
   kubectl config use-context minikube
   ```

3. **Docker image build failures**

   ```bash
   eval $(minikube docker-env)
   docker build -t user-management-service:latest .
   ```

4. **Permission issues**
   ```bash
   sudo chown -R $USER:$USER ~/.kube
   ```

### Debug Mode

For detailed debugging, set the `DEBUG` environment variable:

```bash
DEBUG=1 python scripts/containerManagement/deploy_container.py
```

## Contributing

When modifying these scripts:

1. Follow the existing code style
2. Add proper error handling
3. Include docstrings for all functions
4. Test with different scenarios
5. Update this README if needed

## License

This project is licensed under the MIT License.
