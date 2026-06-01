output "private_ip_address" {
  value = aws_instance.vm.private_ip
}

output "vpc_id" {
  value = aws_vpc.main.id
}

output "account_id" {
  value = data.aws_caller_identity.current.account_id
}
