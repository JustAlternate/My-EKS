resource "aws_ecr_repository" "repositories" {
  for_each = toset(["api", "web-server"])
  name     = each.key
  image_tag_mutability = "MUTABLE"
  force_delete         = true
}

resource "aws_ecr_lifecycle_policy" "repositories_policy" {
  for_each = aws_ecr_repository.repositories

  repository = aws_ecr_repository.repositories[each.key].name

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
      },
      {
        rulePriority = 2
        description  = "Keep only 50 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 50
        }
        action = { 
          type = "expire"
        }
      }
    ]
  })
}
