variable "region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "us-east-1"
}

variable "name" {
  type        = string
  description = "Name to use for all resources if not otherwise specified."
  nullable    = false
}

variable "vpc_name" {
  type        = string
  description = "The name of virutal network."
  nullable    = true
  default     = null
}

variable "subnet_name" {
  type        = string
  description = "Then name of the subnet."
  nullable    = true
  default     = null
}

variable "sg_name" {
  type        = string
  description = "The name of security group."
  nullable    = true
  default     = null
}

variable "vm_name" {
  type        = string
  description = "The name of the VM."
  nullable    = true
  default     = null
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
}

variable "public_subnet_cidr" {
  description = "Public subnet CIDR block"
  type = string
  nullable = true
  default = null
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}
