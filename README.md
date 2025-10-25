# My AWS infra

## Features / TODO

### iac setup
- [x] Store tfstate in S3 bucket
- [x] EKS (AWS managed K8S) + CloudWatch for it
- [x] Managed Node Group multi AZ (only using tg4.small and 2 AZ configured) (arm64, Amazon linux AMI)
- [x] Simple VPC and Security Group for EKS and the Node Group

### Release engineering
- [x] ECR (Managed Registry)
- [x] CI for OpenTofu (lint, format iac code and tf plan and tf apply + comment in the PR)
- [x] CI for our Golang micro-services (lint, format, release, build on arm64 runners and push images on ECR)
- [x] Proper release when rebasing on master & push the release tag on ECR for each services 

### Micro-services and infra setup
- [x] RDS (Simple single RDS that will store a counter and accessible from our micro-services running in EKS) (inside a private subnet with a private Route53 DNS)
- [x] 2 micro-services deployable (one that expose a public web server and send a request to the second one that update the counter in a RDS)
- [x] Deploy manually to EKS (Deployments + Service in simple yaml)
- [x] Healthchecks in the micro-services (liveness, readiness & startup probes)

### Observability
- [ ] Managed Prometheus & Grafana (AMP/AMG)
- [ ] Add Metrics in each micro-services
- [ ] Grafana dashboard for our EKS (CloudWatch)
- [ ] Grafana dashboards for our micro-services
- [ ] Grafana dashboard for our RDS (CloudWatch)
- [ ] Node Exporter for our instances

### Scaling and K8S config
- [ ] Learn, setup and configure HPA (HorizontalPodAutoscaler)
- [ ] Learn and configure PDB (PodDistruptionBudget)
- [ ] Script to simulate traffic and trigger scaling (grafana/k6 ?)
- [ ] Script to randomly kill a pod

### Automatic deployment
- [ ] CD (Argo CD for deploying)
- [ ] CD (Automatic rollback with Argo Rollout)

### SRE stuff
- [ ] Monitor golden signals in Grafana 
- [ ] Define clear SLA / SLI / SLO for our Application feature
- [ ] Monitor our SLI / SLO in Grafana

### Docs
- [ ] Architecture Diagram
- [ ] Create some small ADR for the choices made along the way
- [ ] Finalize README + demo screenshots + demo gif

## Bonus

- [ ] RDS Backup & Restore (Velero ?)
- [ ] Devenv shell to setup dev environment
- [ ] Helm support for apps deployment
- [ ] Karpenter for automatic spot provisionning (replace Node group)
- [ ] Chaos engineering to simulate failing code deployment, kill eks node, slow RDS ? ...
- [ ] Post Mortem
- [ ] Create a Runbook based on the possible failures
- [ ] Simple golang e-BPF tracing
- [ ] Cilium (Circuit breaking ?) 
- [ ] Tempo ? (OpenTelemetry)
- [ ] LinkedIn post lmao

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
# Build them for arm64
docker build --platform linux/arm64 ./apps/api/ -t api -f apps/api/Containerfile
docker build --platform linux/arm64 ./apps/web-server/ -t web-server -f apps/web-server/Containerfile
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

#### Deploy our micro-services to EKS using our stored images in ECR

```
kubectl apply -f ./services/api
kubectl apply -f ./services/web-server
```
