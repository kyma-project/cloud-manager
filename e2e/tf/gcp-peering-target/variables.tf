variable "name" {
  description = "The one name to rule them all. If any of the other variables are not set, this value will be used as the default."
}

variable "network_name" {
  description = "The peering VPC network name."
  nullable = true
  default = null
}
variable "subnetwork_name" {
  description = "Subnet name of the peering target instance."
  nullable = true
  default = null
}
variable "instance_name" {
  description = "The name of the peering target instance."
  nullable = true
  default = null
}
variable "subnet_cidr" {
  description = "The CIDR block of the peering target subnetwork."
  nullable = true
  default = null
}
variable "location" {
  description = "The location of the network."
  nullable = false
}
