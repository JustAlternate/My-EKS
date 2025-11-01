# Data source for AWS account ID
data "aws_caller_identity" "current" {}

# Provider for Kubernetes resources
provider "kubernetes" {
  host                   = aws_eks_cluster.main.endpoint
  cluster_ca_certificate = base64decode(aws_eks_cluster.main.certificate_authority.0.data)
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    args        = ["eks", "get-token", "--cluster-name", aws_eks_cluster.main.name]
    command     = "aws"
  }
}

# Provider for Helm
provider "helm" {
  kubernetes = {
    host                   = aws_eks_cluster.main.endpoint
    cluster_ca_certificate = base64decode(aws_eks_cluster.main.certificate_authority.0.data)
    exec = {
      api_version = "client.authentication.k8s.io/v1beta1"
      args        = ["eks", "get-token", "--cluster-name", aws_eks_cluster.main.name]
      command     = "aws"
    }
  }
}

# EBS CSI Driver addon for EKS
resource "aws_eks_addon" "ebs-csi" {
  cluster_name                = aws_eks_cluster.main.name
  addon_name                  = "aws-ebs-csi-driver"
  addon_version               = "v1.37.0-eksbuild.1"  # Updated to latest stable
  service_account_role_arn    = aws_iam_role.aws_ebs_csi_driver.arn
  resolve_conflicts_on_create = "OVERWRITE"
  resolve_conflicts_on_update = "OVERWRITE"

  depends_on = [
    aws_eks_node_group.main,
    aws_eks_node_group.monitoring
  ]
}

resource "kubernetes_namespace" "monitoring" {
  metadata {
    name = "monitoring"
    
    labels = {
      name = "monitoring"
    }
  }

  depends_on = [
    aws_eks_cluster.main,
    aws_eks_node_group.monitoring
  ]

  timeouts {
    delete = "5m"
  }
}

resource "helm_release" "kube_prometheus_stack" {
  name             = "kube-prometheus-stack"
  repository       = "https://prometheus-community.github.io/helm-charts"
  chart            = "kube-prometheus-stack"
  version          = "65.0.0"
  namespace        = kubernetes_namespace.monitoring.metadata[0].name
  create_namespace = false

  values = [
    file("${path.module}/observability-stack-config/prometheus-stack-values.yaml")
  ]

  # Set Grafana IAM role dynamically
  set = [
    {
      name  = "grafana.serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
      value = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/grafana-cloudwatch-role"
    }
  ]

  atomic          = true
  cleanup_on_fail = true
  wait            = true
  timeout         = 600

  depends_on = [
    aws_eks_addon.ebs-csi,
    aws_eks_node_group.monitoring,
    kubernetes_namespace.monitoring,
    kubernetes_storage_class_v1.gp3
  ]

  lifecycle {
    prevent_destroy = false
  }
}

resource "helm_release" "loki" {
  name             = "loki"
  repository       = "https://grafana.github.io/helm-charts"
  chart            = "loki"
  version          = "6.6.0"
  namespace        = kubernetes_namespace.monitoring.metadata[0].name
  create_namespace = false

  values = [
    file("${path.module}/observability-stack-config/loki-values.yaml")
  ]

  atomic          = true
  cleanup_on_fail = true
  wait            = true  # Changed to true
  timeout         = 600   # 10 minutes

  timeouts = {
    create = "10m"
    update = "10m"
    delete = "10m"
  }

  depends_on = [
    aws_eks_addon.ebs-csi,
    aws_eks_node_group.monitoring,
    kubernetes_namespace.monitoring,
    helm_release.kube_prometheus_stack
  ]
}

resource "helm_release" "promtail" {
  name             = "promtail"
  repository       = "https://grafana.github.io/helm-charts"
  chart            = "promtail"
  version          = "6.16.0"
  namespace        = kubernetes_namespace.monitoring.metadata[0].name
  create_namespace = false

  values = [
    file("${path.module}/observability-stack-config/promtail-values.yaml")
  ]

  atomic          = true
  cleanup_on_fail = true
  wait            = true
  timeout         = 300

  timeouts = {
    create = "10m"
    update = "10m"
    delete = "10m"
  }

  depends_on = [
    helm_release.loki,  # Install after Loki
    kubernetes_namespace.monitoring
  ]
}
