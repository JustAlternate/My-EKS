terraform {
  backend "s3" {
    bucket = "justalternate-multi-project-tfstate-bucket"
    key    = "eks-infra-budget.tfstate"
    region = "eu-west-3"
  }
}
