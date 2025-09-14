# AgentFlow Deployment Guide

This guide covers deploying AgentFlow to production environments, from single-node setups to large-scale Kubernetes clusters.

## Deployment Options

AgentFlow supports multiple deployment patterns:

1. **Single Node** - Docker Compose for small teams/development
2. **Kubernetes** - Production-ready orchestration
3. **Cloud Managed** - Hosted AgentFlow service
4. **Hybrid** - Mix of self-hosted and managed components

## Prerequisites

### System Requirements

#### Minimum (Single Node)
- **CPU**: 4 cores
- **Memory**: 8 GB RAM
- **Storage**: 50 GB SSD
- **Network**: 100 Mbps

#### Recommended (Production)
- **CPU**: 8+ cores per node
- **Memory**: 16+ GB RAM per node  
- **Storage**: 100+ GB SSD per node
- **Network**: 1 Gbps

#### Software Dependencies
- **Docker**: 20.10+
- **Docker Compose**: 2.0+ (for single node)
- **Kubernetes**: 1.25+ (for cluster deployment)
- **kubectl**: Latest version
- **Helm**: 3.0+ (recommended)

### External Services

#### Required
- **PostgreSQL**: 13+ (managed or self-hosted)
- **Redis**: 6+ (for caching and sessions)

#### Optional but Recommended
- **NATS**: 2.9+ (managed or self-hosted)
- **ClickHouse**: 22+ (for analytics)
- **Jaeger**: Latest (for tracing)
- **Prometheus**: Latest (for metrics)

## Single Node Deployment (Docker Compose)

Best for development, testing, and small teams.

### 1. Prepare Environment

```bash
# Create deployment directory
mkdir agentflow-deployment
cd agentflow-deployment

# Download deployment files
curl -L "https://github.com/Siddhant-K-code/agentflow-infrastructure/releases/latest/download/docker-compose-production.yml" -o docker-compose.yml
curl -L "https://github.com/Siddhant-K-code/agentflow-infrastructure/releases/latest/download/env.production" -o .env.example
```

### 2. Configure Environment

```bash
# Copy and edit environment file
cp .env.example .env
nano .env
```

**Required Environment Variables:**

```bash
# Database Configuration
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=agentflow
POSTGRES_USER=agentflow
POSTGRES_PASSWORD=your-secure-password

# Redis Configuration  
REDIS_URL=redis://redis:6379

# NATS Configuration
NATS_URL=nats://nats:4222

# LLM Provider API Keys
OPENAI_API_KEY=sk-your-openai-key
ANTHROPIC_API_KEY=sk-ant-your-anthropic-key

# Security
JWT_SECRET=your-jwt-secret-key
ENCRYPTION_KEY=your-32-character-encryption-key

# Application Configuration
AGENTFLOW_DOMAIN=agentflow.yourdomain.com
AGENTFLOW_PORT=8080
AGENTFLOW_ENV=production

# Optional: Observability
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
PROMETHEUS_URL=http://prometheus:9090
```

### 3. Configure TLS/SSL

Create SSL certificates:

```bash
# Using Let's Encrypt (recommended)
mkdir -p ssl
sudo certbot certonly --standalone \
  -d agentflow.yourdomain.com \
  --cert-path ./ssl/cert.pem \
  --key-path ./ssl/key.pem

# Or create self-signed (development only)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/key.pem \
  -out ssl/cert.pem \
  -subj "/CN=agentflow.yourdomain.com"
```

### 4. Deploy Services

```bash
# Create external networks and volumes
docker network create agentflow-net
docker volume create postgres-data
docker volume create redis-data
docker volume create nats-data

# Deploy all services
docker-compose up -d

# Verify deployment
docker-compose ps
docker-compose logs -f orchestrator
```

### 5. Initialize Database

```bash
# Run database migrations
docker-compose exec orchestrator agentflow migrate up

# Create admin user
docker-compose exec orchestrator agentflow user create \
  --email admin@yourdomain.com \
  --password admin-password \
  --role admin
```

### 6. Verify Deployment

```bash
# Health check
curl -k https://agentflow.yourdomain.com/health

# API test
curl -k https://agentflow.yourdomain.com/api/v1/workflows \
  -H "Authorization: Bearer <token>"
```

## Kubernetes Deployment

Recommended for production environments requiring high availability and scale.

### 1. Prepare Kubernetes Cluster

#### Option A: Managed Kubernetes
- **Google GKE**: Recommended configuration
- **Amazon EKS**: With EBS CSI driver
- **Azure AKS**: With Azure Disk CSI driver
- **DigitalOcean DOKS**: Cost-effective option

#### Option B: Self-Managed
```bash
# Example using kubeadm (Ubuntu 20.04+)
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo kubeadm init --pod-network-cidr=10.244.0.0/16
kubectl apply -f https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml
```

### 2. Install Dependencies

#### Install Helm
```bash
curl https://get.helm.sh/helm-v3.13.0-linux-amd64.tar.gz | tar -zx
sudo mv linux-amd64/helm /usr/local/bin/helm
```

#### Add Helm Repositories
```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add jetstack https://charts.jetstack.io
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
```

### 3. Install Infrastructure Components

#### Namespace
```bash
kubectl create namespace agentflow
kubectl label namespace agentflow istio-injection=enabled
```

#### PostgreSQL
```bash
helm install postgresql bitnami/postgresql \
  --namespace agentflow \
  --set auth.postgresPassword=your-postgres-password \
  --set auth.database=agentflow \
  --set primary.persistence.size=100Gi \
  --set primary.resources.requests.memory=2Gi \
  --set primary.resources.requests.cpu=1000m \
  --set primary.resources.limits.memory=4Gi \
  --set primary.resources.limits.cpu=2000m
```

#### Redis
```bash
helm install redis bitnami/redis \
  --namespace agentflow \
  --set auth.password=your-redis-password \
  --set master.persistence.size=20Gi \
  --set replica.replicaCount=2
```

#### NATS
```bash
helm install nats nats/nats \
  --namespace agentflow \
  --set nats.jetstream.enabled=true \
  --set nats.jetstream.fileStore.pvc.size=50Gi \
  --set cluster.enabled=true \
  --set cluster.replicas=3
```

#### Ingress Controller
```bash
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=LoadBalancer
```

#### Cert Manager (for TLS)
```bash
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true
```

### 4. Deploy AgentFlow

#### Download Kubernetes Manifests
```bash
curl -L "https://github.com/Siddhant-K-code/agentflow-infrastructure/releases/latest/download/kubernetes-manifests.tar.gz" | tar -zx
cd kubernetes-manifests/
```

#### Configure Secrets
```bash
# Create secrets
kubectl create secret generic agentflow-secrets \
  --namespace agentflow \
  --from-literal=postgres-password=your-postgres-password \
  --from-literal=redis-password=your-redis-password \
  --from-literal=jwt-secret=your-jwt-secret \
  --from-literal=encryption-key=your-32-char-key \
  --from-literal=openai-api-key=sk-your-key \
  --from-literal=anthropic-api-key=sk-ant-your-key
```

#### Deploy Services
```bash
# Apply manifests
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secrets.yaml
kubectl apply -f orchestrator/
kubectl apply -f runtime/
kubectl apply -f llm-router/
kubectl apply -f ingress/

# Wait for deployment
kubectl wait --for=condition=available --timeout=300s \
  deployment/orchestrator deployment/runtime deployment/llm-router \
  -n agentflow
```

#### Configure Ingress and TLS
```bash
# Create ClusterIssuer for Let's Encrypt
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@yourdomain.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF

# Create Ingress
cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: agentflow-ingress
  namespace: agentflow
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - api.agentflow.yourdomain.com
    secretName: agentflow-tls
  rules:
  - host: api.agentflow.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: orchestrator
            port:
              number: 8080
EOF
```

### 5. Configure Auto-Scaling

#### Horizontal Pod Autoscaler
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: orchestrator-hpa
  namespace: agentflow
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: orchestrator
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

#### Vertical Pod Autoscaler (Optional)
```bash
# Install VPA
kubectl apply -f https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/deploy/vpa-v1-crd-gen.yaml
kubectl apply -f https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/deploy/vpa-rbac.yaml
kubectl apply -f https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/deploy/vpa-deployment.yaml
```

### 6. Setup Monitoring

#### Prometheus and Grafana
```bash
# Add monitoring repository
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts

# Install Prometheus stack
helm install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set grafana.adminPassword=admin-password \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=50Gi
```

#### Jaeger for Distributed Tracing
```bash
# Install Jaeger operator
kubectl create namespace observability
kubectl apply -f https://github.com/jaegertracing/jaeger-operator/releases/latest/download/jaeger-operator.yaml -n observability

# Deploy Jaeger instance
cat <<EOF | kubectl apply -f -
apiVersion: jaegertracing.io/v1
kind: Jaeger
metadata:
  name: agentflow-jaeger
  namespace: agentflow
spec:
  strategy: production
  storage:
    type: elasticsearch
    elasticsearch:
      nodeCount: 3
      storage:
        size: 100Gi
      redundancyPolicy: SingleRedundancy
EOF
```

## Cloud Managed Deployment

For organizations preferring managed services.

### AWS EKS with Managed Services

#### Infrastructure Setup
```bash
# Create EKS cluster
eksctl create cluster \
  --name agentflow \
  --region us-west-2 \
  --nodegroup-name standard-workers \
  --node-type m5.large \
  --nodes 3 \
  --nodes-min 1 \
  --nodes-max 10 \
  --managed

# Install AWS Load Balancer Controller
kubectl apply -k "github.com/aws/eks-charts/stable/aws-load-balancer-controller//crds?ref=master"
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  --set clusterName=agentflow \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller \
  -n kube-system
```

#### Managed Services
```bash
# Create RDS PostgreSQL
aws rds create-db-instance \
  --db-instance-identifier agentflow-postgres \
  --db-instance-class db.t3.medium \
  --engine postgres \
  --engine-version 14.9 \
  --allocated-storage 100 \
  --storage-type gp2 \
  --storage-encrypted \
  --master-username agentflow \
  --master-user-password your-password \
  --vpc-security-group-ids sg-xxxxxxxx \
  --db-subnet-group-name agentflow-subnet-group

# Create ElastiCache Redis
aws elasticache create-replication-group \
  --replication-group-id agentflow-redis \
  --description "AgentFlow Redis Cluster" \
  --node-type cache.t3.medium \
  --num-cache-clusters 2 \
  --engine redis \
  --engine-version 6.2 \
  --cache-parameter-group-name default.redis6.x \
  --cache-subnet-group-name agentflow-cache-subnet-group
```

### Google GKE with Cloud Services

#### GKE Autopilot Cluster
```bash
# Create Autopilot cluster
gcloud container clusters create-auto agentflow \
  --region=us-central1 \
  --release-channel=regular

# Get credentials
gcloud container clusters get-credentials agentflow --region=us-central1
```

#### Cloud SQL and Memorystore
```bash
# Create Cloud SQL PostgreSQL
gcloud sql instances create agentflow-postgres \
  --database-version=POSTGRES_14 \
  --tier=db-standard-2 \
  --region=us-central1 \
  --storage-size=100GB \
  --storage-type=SSD

# Create Memorystore Redis
gcloud redis instances create agentflow-redis \
  --size=5 \
  --region=us-central1 \
  --redis-version=redis_6_x
```

## Security Configuration

### Network Security

#### Network Policies
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: agentflow-network-policy
  namespace: agentflow
spec:
  podSelector:
    matchLabels:
      app: orchestrator
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: agentflow
    ports:
    - protocol: TCP
      port: 5432  # PostgreSQL
    - protocol: TCP
      port: 6379  # Redis
    - protocol: TCP
      port: 4222  # NATS
```

#### Pod Security Policies
```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: agentflow-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
```

### RBAC Configuration

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: agentflow
  name: agentflow-operator
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: agentflow-operator-binding
  namespace: agentflow
subjects:
- kind: ServiceAccount
  name: agentflow-operator
  namespace: agentflow
roleRef:
  kind: Role
  name: agentflow-operator
  apiGroup: rbac.authorization.k8s.io
```

## Performance Tuning

### Database Optimization

#### PostgreSQL Configuration
```sql
-- postgresql.conf optimizations
shared_buffers = 1GB
effective_cache_size = 3GB
maintenance_work_mem = 256MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 4MB
min_wal_size = 1GB
max_wal_size = 4GB
max_connections = 200
```

#### Database Indexing
```sql
-- Performance indexes for AgentFlow
CREATE INDEX CONCURRENTLY idx_workflows_status ON workflows(status);
CREATE INDEX CONCURRENTLY idx_executions_workflow_status ON workflow_executions(workflow_id, status);
CREATE INDEX CONCURRENTLY idx_executions_created_at ON workflow_executions(created_at DESC);
CREATE INDEX CONCURRENTLY idx_agent_executions_workflow ON agent_executions(workflow_execution_id);
CREATE INDEX CONCURRENTLY idx_metrics_timestamp ON execution_metrics(timestamp DESC);
```

### Application Performance

#### JVM Tuning (for Java components)
```bash
JAVA_OPTS="-Xms2g -Xmx4g -XX:+UseG1GC -XX:MaxGCPauseMillis=200 -XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap"
```

#### Go Runtime Tuning
```bash
# Environment variables for Go services
GOMAXPROCS=4
GOGC=100
GOMEMLIMIT=2GiB
```

### Resource Limits

```yaml
resources:
  requests:
    memory: "1Gi"
    cpu: "500m"
  limits:
    memory: "2Gi" 
    cpu: "1000m"
```

## Backup and Disaster Recovery

### Database Backup

#### Automated Backup Script
```bash
#!/bin/bash
# backup-postgres.sh

POSTGRES_HOST="your-postgres-host"
POSTGRES_DB="agentflow"
POSTGRES_USER="agentflow"
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup
pg_dump -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB \
  --no-password --format=custom \
  --file="$BACKUP_DIR/agentflow_backup_$DATE.sql"

# Compress backup
gzip "$BACKUP_DIR/agentflow_backup_$DATE.sql"

# Upload to cloud storage (optional)
aws s3 cp "$BACKUP_DIR/agentflow_backup_$DATE.sql.gz" \
  s3://your-backup-bucket/postgres/

# Clean up old backups (keep 7 days)
find $BACKUP_DIR -name "agentflow_backup_*.sql.gz" -mtime +7 -delete
```

#### Kubernetes CronJob for Backups
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: agentflow
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: postgres-backup
            image: postgres:14
            command:
            - /bin/bash
            - -c
            - |
              pg_dump -h postgres -U agentflow -d agentflow \
                --no-password --format=custom \
                --file=/backup/agentflow_$(date +%Y%m%d_%H%M%S).sql
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-secret
                  key: password
            volumeMounts:
            - name: backup-storage
              mountPath: /backup
          volumes:
          - name: backup-storage
            persistentVolumeClaim:
              claimName: backup-pvc
          restartPolicy: OnFailure
```

### Disaster Recovery Plan

#### Recovery Time Objectives (RTO)
- **Critical Services**: 15 minutes
- **Full System**: 1 hour
- **Historical Data**: 4 hours

#### Recovery Point Objectives (RPO)
- **Workflow State**: 5 minutes
- **Execution Data**: 15 minutes
- **Metrics Data**: 1 hour

#### Recovery Procedures

1. **Service Failure**
```bash
# Check service health
kubectl get pods -n agentflow
kubectl describe pod <failing-pod> -n agentflow

# Restart service
kubectl rollout restart deployment/orchestrator -n agentflow
```

2. **Database Failure**
```bash
# Restore from backup
pg_restore -h new-postgres-host -U agentflow -d agentflow \
  --clean --if-exists \
  /backups/agentflow_backup_latest.sql
```

3. **Complete Cluster Failure**
```bash
# Restore cluster from infrastructure as code
terraform apply -var="environment=production"

# Restore data from backups
kubectl apply -f kubernetes-manifests/
# Wait for pods to be ready
# Restore database
# Verify service functionality
```

## Monitoring and Alerting

### Key Metrics to Monitor

#### System Metrics
- CPU and memory usage per pod
- Network I/O and latency
- Disk usage and I/O
- Pod restart frequency

#### Application Metrics  
- Workflow execution success rate
- Average execution duration
- Queue depth and processing rate
- LLM API response times and error rates

#### Business Metrics
- Active workflows count
- Daily execution volume
- Cost per execution
- User activity and growth

### Alerting Rules

```yaml
# prometheus-alerts.yaml
groups:
- name: agentflow.rules
  rules:
  - alert: HighErrorRate
    expr: rate(agentflow_executions_failed_total[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High execution failure rate"
      description: "Execution failure rate is {{ $value }} for the past 5 minutes"

  - alert: DatabaseConnectionFailure
    expr: up{job="postgres"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Database connection failed"
      description: "PostgreSQL database is not responding"

  - alert: HighMemoryUsage
    expr: container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.9
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High memory usage detected"
      description: "Container {{ $labels.container }} is using {{ $value }}% of available memory"
```

## Troubleshooting

### Common Issues

#### 1. Pod Startup Failures
```bash
# Check pod status
kubectl get pods -n agentflow

# View pod logs
kubectl logs -f <pod-name> -n agentflow

# Describe pod for events
kubectl describe pod <pod-name> -n agentflow

# Check resource constraints
kubectl top pods -n agentflow
```

#### 2. Database Connection Issues
```bash
# Test database connectivity
kubectl run -it --rm debug --image=postgres:14 --restart=Never -- \
  psql -h postgres -U agentflow -d agentflow

# Check database performance
kubectl exec -it postgres-0 -n agentflow -- \
  psql -U agentflow -d agentflow -c "SELECT * FROM pg_stat_activity;"
```

#### 3. High Latency Issues
```bash
# Check network policies
kubectl get networkpolicies -n agentflow

# Test service connectivity
kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -- \
  curl -v http://orchestrator:8080/health

# Check ingress configuration
kubectl get ingress -n agentflow
kubectl describe ingress agentflow-ingress -n agentflow
```

## Maintenance

### Regular Maintenance Tasks

#### Weekly
- Review monitoring dashboards and alerts
- Check backup integrity and restoration procedures
- Update non-critical dependencies
- Review security logs and access patterns

#### Monthly
- Performance review and optimization
- Capacity planning and scaling review
- Security patch updates
- Database maintenance and optimization

#### Quarterly
- Disaster recovery testing
- Security audit and penetration testing
- Infrastructure cost optimization
- Technology stack updates and migrations

### Update Procedures

#### Rolling Updates
```bash
# Update orchestrator
kubectl set image deployment/orchestrator \
  orchestrator=agentflow/orchestrator:v1.1.0 \
  -n agentflow

# Check rollout status
kubectl rollout status deployment/orchestrator -n agentflow

# Rollback if needed
kubectl rollout undo deployment/orchestrator -n agentflow
```

This comprehensive deployment guide provides everything needed to run AgentFlow reliably in production environments, from small teams to enterprise scale.