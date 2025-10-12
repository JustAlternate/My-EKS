# My AWS infra

## Features / TODO

- [x] tfstate Backend bucket S3
- [x] EKS (AWS managed K8S)
- [x] Managed Node Group multi AZ (only using tg4.small and 2 dynamic AZ configured)
- [x] CloudWatch
- [ ] 2 Apps deployable (one that expose a public web-server with a button to send a request to the second one that update the counter in the RDS)
- [ ] RDS (postgresql that will store a counter)
- [ ] Helm support for apps deployment
- [ ] ECR (Managed Registry)
- [ ] CI (lint, format TF & lint, format, build, push Containerfile)
- [ ] CD (Argo CD for deploying images and infrastructure)
- [ ] Karpenter for automatic pods provisionning (replace Node group)
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
ECR_URL="https://$(aws sts get-caller-identity --query 'Account' --output text).dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com"
aws ecr get-login-password --region $AWS_DEFAULT_REGION | docker login -u AWS --password-stdin $ECR_URL
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

```
TODO curl
```


### Deploy our own apps

#### Create repositories if not exist

Add your repository for each image you want to build and store
```
nvim iac/storage/ecr.tf
```

#### Build and Deploy

```
ECR_URL=$(aws sts get-caller-identity --query 'Account' --output text).dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com

# Build them
docker build ./apps/api/ -t api -f apps/api/Containerfile
docker build ./apps/web-server/ -t web-server -f apps/web-server/Containerfile

# Tag them
docker tag api:latest $ECR_URL/api:latest
docker tag web-server:latest $ECR_URL/web-server:latest

# Push them
docker push $ECR_URL/api:latest
docker push $ECR_URL/web-server:latest
```
