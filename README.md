# My AWS infra

## Features / TODO

- [x] tfstate Backend bucket S3
- [x] EKS (AWS managed K8S)
- [x] Managed Node Group multi AZ (only using tg4.small and 2 dynamic AZ configured)
- [x] CloudWatch
- [ ] Helm support for apps deployment
- [ ] RDS (postgresql that will store a counter)
- [ ] 2 Apps deployable (one that expose a public web-server with a button to send a request to the second one that update the counter in the RDS)
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
tofu init

tofu plan

tofu apply
```

Get access to the cluster through kubectl on remote machine
```
aws eks update-kubeconfig --region $AWS_DEFAULT_REGION --name justalternate-eks-cluster 
```

Deploy with kubectl directly
```
kubectl run test-pod --image=nginx --restart=Never
```
