# IAM Users
resource "aws_iam_user" "github_actions_deploy" {
  name = "github-actions-deploy"
  path = "/"

  tags = {
    Name        = "github-actions-deploy"
    Environment = var.stack_name
    Purpose     = "GitHub Actions Deployment"
  }
}

resource "aws_iam_user" "john_mason" {
  name = "JohnMason"
  path = "/"

  tags = {
    Name        = "JohnMason"
    Environment = var.stack_name
    Purpose     = "Administrator"
  }
}

# IAM User Group - Cloud_Watch
resource "aws_iam_group" "cloud_watch" {
  name = "Cloud_Watch"
  path = "/"
}

# Attach CloudWatch policies to the Cloud_Watch group
resource "aws_iam_group_policy_attachment" "cloud_watch_read_only" {
  group      = aws_iam_group.cloud_watch.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchReadOnlyAccess"
}

resource "aws_iam_group_policy_attachment" "cloud_watch_logs_read_only" {
  group      = aws_iam_group.cloud_watch.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsReadOnlyAccess"
}

# Attach Administrator Access to both users
resource "aws_iam_user_policy_attachment" "john_mason_admin" {
  user       = aws_iam_user.john_mason.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}

resource "aws_iam_user_policy_attachment" "github_actions_admin" {
  user       = aws_iam_user.github_actions_deploy.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}

# Add JohnMason to Cloud_Watch group
resource "aws_iam_user_group_membership" "john_mason_cloud_watch" {
  user = aws_iam_user.john_mason.name
  groups = [
    aws_iam_group.cloud_watch.name,
  ]
}
