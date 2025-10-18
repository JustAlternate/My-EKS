# My AWS infra

## Features / TODO

- [x] State tfstate Backend bucket S3
- [x] EKS (AWS managed K8S)
- [x] CloudWatch for K8S
- [x] Managed Node Group multi AZ (only using tg4.small and 2 AZ configured)
- [x] Simple VPC and Security Group for EKS and the Node Group
- [x] ECR (Managed Registry)
- [x] CI for OpenTofu (lint, format iac code and tf plan and tf apply + comment in the PR)
- [x] CI for our Golang Micro-services (lint, format, release, build, push Containerfile)
- [ ] RDS (simple postgresql that will store a counter)
- [ ] 2 micro-services deployable (one that expose a public web server and send a request to the second one that update the counter in a RDS)
- [ ] Setup and configure HPA & PDB (HorizontalPodAutoscaler & PodDistruptionBudget)
- [ ] Healthchecks in the micro-services (liveness & readiness)
- [ ] Prometheus & Grafana on a static EC2 separated from the cluster (avoiding SPOF for the Observability Stack ? Or maybe go with AMP/AMG)
- [ ] Craft great dashboard for the RDS, EKS and the Apps
- [ ] Add Metrics in each micro-services
- [ ] CD (Argo CD for deploying)
- [ ] Architecture Diagram
- [ ] CD (Automatic rollback with Argo Rollout)
- [ ] Define clear SLA / SLI / SLO for our Application feature
- [ ] Monitor the SLI/SLO using a Grafana dashboard
- [ ] Script to simulate traffic and trigger scaling (grafana/k6 ?)
- [ ] Restauration RDS (Velero ?)

## Bonus

- [ ] Devenv shell to setup dev environment
- [ ] Helm support for apps deployment
- [ ] Karpenter for automatic spot provisionning (replace Node group)
- [ ] Chaos engineering to simulate failing code deployment, kill eks node, slow RDS ? ...
- [ ] Create a Post Mortem
- [ ] Create a Runbook based on the possible failures
- [ ] Simple e-BPF tracing
- [ ] Cilium (Circuit breaking ?) 
- [ ] Loki & Tempo (OpenTelemetry)

## Init

```
cp env.dist .env
```

```
tofu -chdir=iac init

tofu -chdir=iac plan 

tofu -chdir=iac apply
```

Get access to the cluster through kubectl on remote machine
```
aws eks update-kubeconfig --region $AWS_DEFAULT_REGION --name justalternate-eks-cluster 
```

Login to the ECR in case you want to manually push image to it
```
AWS_FULL_ECR_URL="https://$(aws sts get-caller-identity --query 'Account' --output text).dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com"
aws ecr get-login-password --region $AWS_DEFAULT_REGION | docker login -u AWS --password-stdin $AWS_FULL_ECR_URL
```

## Deploy some apps

### Deploy public containers

Deploy with kubectl directly
```
kubectl run test-pod --image=nginx --restart=Never
```

Or 

Using yaml definition :

```
kubectl apply -f services/nginx-test-hello-world
```
This will create a Deployment of nginx with 2 replicas as well as a Load Balancer to access it

### Deploy our own apps

#### Create ECR repository if not exist

Add your repository for each app you want to build and store
```
nvim iac/storage/ecr.tf
```

#### Build and push to ecr locally

```
AWS_ACCOUNT=$(aws sts get-caller-identity --query 'Account' --output text)
AWS_ECR_URL=$AWS_ACCOUNT.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com
```

```
# Build them
docker build ./apps/api/ -t api -f apps/api/Containerfile
docker build ./apps/web-server/ -t web-server -f apps/web-server/Containerfile
```

```
# Tag them
docker tag api:latest $AWS_ECR_URL/api:latest
docker tag web-server:latest $AWS_ECR_URL/web-server:latest
```

```
# Push them
docker push $AWS_ECR_URL/api:latest
docker push $AWS_ECR_URL/web-server:latest
```

Or 

Copy `.github/workflows/build-push-dev.yml` and create one for your service

#### Deploy our images from ECR

```
todo
```
