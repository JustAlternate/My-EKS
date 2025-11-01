# EKS Node Group
resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "${var.cluster_name}-nodes"
  node_role_arn   = aws_iam_role.eks_nodes.arn
  subnet_ids      = module.vpc.private_subnets

  scaling_config {
    desired_size = var.node_desired_size
    max_size     = var.node_max_size
    min_size     = var.node_min_size
  }

  update_config {
    max_unavailable = 1
  }

  instance_types = var.node_instance_types
  ami_type       = "AL2023_ARM_64_STANDARD" # ARM AMI for t4g instances
  # release_version will be automatically set to match the cluster version

  depends_on = [
    aws_iam_role_policy_attachment.eks_nodes_AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.eks_nodes_AmazonEKS_CNI_Policy,
    aws_iam_role_policy_attachment.eks_nodes_AmazonEC2ContainerRegistryReadOnly,
  ]

  tags = {
    Name = "${var.cluster_name}-node-group"
  }

  # Increase timeout for node group creation (can take 10-20 minutes)
  timeouts {
    create = "40m"
    update = "40m"
    delete = "40m"
  }

  # Force node group update when cluster version changes
  lifecycle {
    ignore_changes = [release_version]
  }
}

# Monitoring Node Group (dedicated for observability)
resource "aws_eks_node_group" "monitoring" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "${var.cluster_name}-monitoring-nodes"
  node_role_arn   = aws_iam_role.eks_nodes.arn  # Can reuse same role
  subnet_ids      = module.vpc.private_subnets

  scaling_config {
    desired_size = 1
    max_size     = 2
    min_size     = 1
  }

  update_config {
    max_unavailable = 1
  }

  instance_types = ["t4g.medium"]  # ARM instance, enough for Prometheus/Grafana
  ami_type       = "AL2023_ARM_64_STANDARD"

  labels = {
    workload = "monitoring"
    role     = "observability"
  }

  taint {
    key    = "monitoring"
    value  = "true"
    effect = "NO_SCHEDULE"
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_nodes_AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.eks_nodes_AmazonEKS_CNI_Policy,
    aws_iam_role_policy_attachment.eks_nodes_AmazonEC2ContainerRegistryReadOnly,
  ]

  tags = {
    Name = "${var.cluster_name}-monitoring-node-group"
    Type = "monitoring"
  }

  timeouts {
    create = "40m"
    update = "40m"
    delete = "40m"
  }

  lifecycle {
    ignore_changes = [release_version]
  }
}
