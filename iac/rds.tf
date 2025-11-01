# DB Subnet Group
resource "aws_db_subnet_group" "postgres" {
  name       = "${var.cluster_name}-postgres-subnet-group"
  subnet_ids = module.vpc.private_subnets

  tags = {
    Name = "${var.cluster_name}-postgres-subnet-group"
  }
}

# Security Group for RDS
resource "aws_security_group" "postgres" {
  name_prefix = "${var.cluster_name}-rds-postgres-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [
      aws_security_group.eks_nodes.id,
      aws_eks_cluster.main.vpc_config[0].cluster_security_group_id # Allow this sg that is automatically created by AWS, else it will not work.
    ]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.cluster_name}-rds-postgres-sg"
  }
}

resource "aws_db_instance" "postgres" {
  identifier        = "${var.cluster_name}-postgres"
  allocated_storage = 5
  db_name           = "my_postgres"
  engine            = "postgres" # We will not use Aurora so we can create our own backup and restoration strategie
  engine_version    = "17.6"
  instance_class    = "db.t4g.micro"
  username          = "username"
  password          = "password"

  skip_final_snapshot = true
  publicly_accessible = false
  deletion_protection = false
  final_snapshot_identifier = null
  delete_automated_backups  = true

  db_subnet_group_name    = aws_db_subnet_group.postgres.name
  vpc_security_group_ids  = [aws_security_group.postgres.id]
  multi_az                = false
  backup_retention_period = 7
  apply_immediately       = true

  tags = {
    Name = "${var.cluster_name}-postgres"
  }
}

# Private hosted zone for internal DNS
resource "aws_route53_zone" "private" {
  name = "${var.cluster_name}.internal"

  vpc {
    vpc_id = module.vpc.vpc_id
  }

  tags = {
    Name = "${var.cluster_name}-private-zone"
  }
}

# CNAME record pointing to RDS endpoint
resource "aws_route53_record" "postgres" {
  zone_id = aws_route53_zone.private.zone_id
  name    = "postgres.${var.cluster_name}.internal"
  type    = "CNAME"
  ttl     = 300
  records = [aws_db_instance.postgres.address]
}
