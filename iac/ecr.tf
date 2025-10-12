resource "aws_ecr_repository" "repositories" {
  for_each = toset( ["api", "web-server"] )
  name     = each.key
  image_tag_mutability = "IMMUTABLE"
}
