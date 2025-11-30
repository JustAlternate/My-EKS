# IAM for ESO
resource "aws_iam_policy" "eso_shared" {
  name        = "eso-shared-secret-policy"
  description = "ESO ClusterSecretStore policy"
  policy      = data.aws_iam_policy_document.eso_policy.json
}

# Policy to apply to ESO
data "aws_iam_policy_document" "eso_policy" {
  statement {
    actions = [
      "secretsmanager:GetResourcePolicy",
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
      "secretsmanager:ListSecrets",
      "secretsmanager:ListSecretVersionIds",
    ]
    resources = [aws_secretsmanager_secret.rds-secret.arn]
  }
}

# IAM role for service account to access sercrets manager (IRSA)
resource "aws_iam_role" "eso_service_account_role" {
  name = "${var.cluster_name}-eso-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = aws_iam_openid_connect_provider.eks_oidc.arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            # ESO’s default service account
            "${replace(aws_eks_cluster.main.identity[0].oidc[0].issuer, "https://", "")}:sub" = "system:serviceaccount:external-secrets:external-secrets"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "eso_policy_attach" {
  role       = aws_iam_role.eso_service_account_role.name
  policy_arn = aws_iam_policy.eso_shared.arn
}
