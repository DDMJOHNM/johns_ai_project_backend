output "instance_id" {
  description = "EC2 Instance ID"
  value       = aws_instance.backend.id
}

output "public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_instance.backend.public_ip
}

output "public_dns" {
  description = "Public DNS name of the EC2 instance"
  value       = aws_instance.backend.public_dns
}

output "backend_url" {
  description = "Backend server URL"
  value       = "http://${aws_instance.backend.public_dns}:8080"
}

output "security_group_id" {
  description = "Security Group ID"
  value       = aws_security_group.ec2_sg.id
}

output "iam_role_name" {
  description = "IAM Role name for the EC2 instance"
  value       = aws_iam_role.ec2_role.name
}

output "iam_role_arn" {
  description = "IAM Role ARN for the EC2 instance"
  value       = aws_iam_role.ec2_role.arn
}

