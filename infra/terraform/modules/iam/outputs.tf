output "github_actions_user_arn" {
  description = "ARN of the github-actions-deploy user"
  value       = aws_iam_user.github_actions_deploy.arn
}

output "john_mason_user_arn" {
  description = "ARN of the JohnMason user"
  value       = aws_iam_user.john_mason.arn
}

output "cloud_watch_group_name" {
  description = "Name of the Cloud_Watch IAM group"
  value       = aws_iam_group.cloud_watch.name
}

output "cloud_watch_group_arn" {
  description = "ARN of the Cloud_Watch IAM group"
  value       = aws_iam_group.cloud_watch.arn
}

