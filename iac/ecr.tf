resource "aws_ecr_repository" "repositories" {
  for_each = toset(["api", "web-server"])
  name     = each.key
  # image_tag_mutability = "IMMUTABLE"
  image_tag_mutability = "MUTABLE" //TODO: fix this to IMMUTABLE once you have proper release tags image build and pushing to ECR (ie = a CI/CD)
  force_delete         = true
}
