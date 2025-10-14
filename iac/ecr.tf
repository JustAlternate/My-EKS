resource "aws_ecr_repository" "prod_repositories" {
  for_each = toset(["api", "web-server"])
  name     = "${each.key}-prod"
  image_tag_mutability = "IMMUTABLE"
  force_delete         = true
}

resource "aws_ecr_lifecycle_policy" "prod_repositories_policy" {
  for_each = aws_ecr_repository.prod_repositories

  repository = aws_ecr_repository.prod_repositories[each.key].name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep production images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 50
        }
        action = { type = "expire" }
      }
    ]
  })
}

resource "aws_ecr_repository" "dev_repositories" {
  for_each = toset(["api", "web-server"])
  name     = "${each.key}-dev"
  image_tag_mutability = "MUTABLE"
  force_delete         = true
}

resource "aws_ecr_lifecycle_policy" "dev_repositories_policy" {
  for_each = aws_ecr_repository.dev_repositories

  repository = aws_ecr_repository.dev_repositories[each.key].name

  policy = jsonencode({
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
