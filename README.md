# My AWS infra

## Features / TODO

- [x] tfstate Backend bucket S3
- [x] EKS (AWS managed K8S)
- [x] Managed Node Group multi AZ (only using tg4.small and 2 dynamic AZ configured)
- [x] ECR (Managed Registry)
- [x] CloudWatch
- [x] CI (lint, format iac code and tf plan and tf apply)
- [ ] CI (lint, format, build, push Containerfile)
- [ ] 2 Apps deployable (one that expose a public web-server and send a request to the second one that update the counter in a RDS)
- [ ] add healthchecks to the apps (liveness & readiness)
- [ ] RDS (postgresql that will store a counter)
- [ ] CD (Argo CD for deploying images)
- [ ] CD (Automatic rollback on liveness failure)
- [ ] Devenv shell to setup dev environment
- [ ] Helm support for apps deployment
- [ ] Karpenter for automatic spot provisionning (replace Node group)
- [ ] Grafana for visualizing prometheus and CloudWatch metrics

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

Login to the ECR to manually push image to it
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

#### Create repositories if not exist

Add your repository for each app you want to build and store
```
nvim iac/storage/ecr.tf
```

#### Build and push to ecr

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
