# Data source for availability zones
data "aws_availability_zones" "available" {
  state = "available"
}

# VPC Module
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "6.4.0"

  name = "${var.cluster_name}-vpc"
  cidr = var.vpc_cidr

  # Use first 2 availability zones dynamically
  azs = slice(data.aws_availability_zones.available.names, 0, 2)

  # Calculate subnets dynamically using cidrsubnet function
  public_subnets = [
    cidrsubnet(var.vpc_cidr, 8, 1), # 10.0.1.0/24
    cidrsubnet(var.vpc_cidr, 8, 2)  # 10.0.2.0/24
  ]

  private_subnets = [
    cidrsubnet(var.vpc_cidr, 8, 10), # 10.0.10.0/24
    cidrsubnet(var.vpc_cidr, 8, 20)  # 10.0.20.0/24
  ]

  # Use single NAT gateway for cost optimization
  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_vpn_gateway   = false
  enable_dns_hostnames = true
  enable_dns_support   = true

  # EKS specific tags
  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
  }

  tags = {
    Name        = "${var.cluster_name}-vpc"
    Environment = "production"
  }
}
