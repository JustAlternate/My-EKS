resource "aws_ecr_repository" "repositories" {
  for_each = toset(["api", "web-server"])
  name     = "${each.key}-prod"
  image_tag_mutability = "IMMUTABLE"
  force_delete         = true

  lifecycle_policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep production images"
        selection = {
          tagStatus   = "tagged"
          countType   = "imageCountMoreThan"
          countNumber = 50
        }
        action = { type = "expire" }
      }
    ]
  })
}


resource "aws_ecr_repository" "repositories" {
  for_each = toset(["api", "web-server"])
  name     = "${each.key}-dev"
  image_tag_mutability = "MUTABLE"
  force_delete         = true
  lifecycle_policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Expire untagged images older than 1 day"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 1
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}
