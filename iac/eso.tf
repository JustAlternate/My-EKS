module "eso_shared_secret_role" {
  source = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  trusted_role_arns = []
  create_role       = true
  role_name         = "eso-shared-secret-assumable-role"
  custom_role_policy_arns = [
    aws_iam_policy.eso_shared.arn
  ]
  number_of_custom_role_policy_arns = 1
  role_requires_mfa                 = false
}

resource "aws_iam_policy" "eso_shared" {
  name        = "eso-shared-secret-policy"
  description = "ESO ClusterSecretStore policy"
  policy      = data.aws_iam_policy_document.eso_shared_secret_policy.json
}

data "aws_iam_policy_document" "eso_shared_secret_policy" {
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

# Policy to apply to a secret to allow eso to read it
data "aws_iam_policy_document" "eso" {
  statement {
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = [module.eso_shared_secret_role.iam_role_arn]
    }
    actions = [
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
    ]
    resources = ["*"]
  }
}
