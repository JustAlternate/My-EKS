#!/usr/bin/env bash
set -e

AWS_FULL_ECR_URL="https://$(aws sts get-caller-identity --query 'Account' --output text).dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com"
aws ecr get-login-password --region $AWS_DEFAULT_REGION | docker login -u AWS --password-stdin $AWS_FULL_ECR_URL

AWS_ACCOUNT=$(aws sts get-caller-identity --query 'Account' --output text)
AWS_ECR_URL=$AWS_ACCOUNT.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com

docker build --platform linux/arm64 ./apps/api/ -t api -f apps/api/Containerfile
docker build --platform linux/arm64 ./apps/web-server/ -t web-server -f apps/web-server/Containerfile

docker tag api:latest $AWS_ECR_URL/api:latest
docker tag web-server:latest $AWS_ECR_URL/web-server:latest

docker push $AWS_ECR_URL/api:latest
docker push $AWS_ECR_URL/web-server:latest
