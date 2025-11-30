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
    from_port = 5432
    to_port   = 5432
    protocol  = "tcp"
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

resource "random_password" "password" {
  length  = 16
  special = false
}

resource "aws_secretsmanager_secret" "rds-secret" {
  name                           = "my_postgres"
  force_overwrite_replica_secret = true
  recovery_window_in_days        = 0
}

resource "aws_secretsmanager_secret_version" "rds-secret-version" {
  secret_id = aws_secretsmanager_secret.rds-secret.id
  secret_string = jsonencode({
    password = random_password.password.result
    username = aws_db_instance.postgres.username
    host     = aws_db_instance.postgres.address
    port     = aws_db_instance.postgres.port
    dbname   = aws_db_instance.postgres.db_name
  })
}

resource "aws_db_instance" "postgres" {
  identifier        = "${var.cluster_name}-postgres"
  allocated_storage = 5
  db_name           = "my_postgres"
  engine            = "postgres" # We will not use Aurora so we can create our own backup and restoration strategie
  engine_version    = "17.6"
  instance_class    = "db.t4g.micro"
  username          = "master"
  password          = random_password.password.result

  skip_final_snapshot       = true
  publicly_accessible       = false
  deletion_protection       = false
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
