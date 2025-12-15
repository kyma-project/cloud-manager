
variable "location" {
  type = string
  description = "The Azure Region where the Resource Group should exist."
  nullable = false
}


variable "virtual_network_address_space" {
  type = string
  description = "The address space that is used the virtual network."
  nullable = false
}

variable "subnet_address_prefix" {
  type = string
  description = "The address prefixes to use for the subnet. Defaults to the `virtual_network_address_space` input value."
  nullable = true
  default = null
}

variable "name" {
  type = string
  description = "Name to use for all resources if not otherwise specified."
  nullable = false
}

variable "resource_group_name" {
  type = string
  description = "The Name which should be used for this Resource Group. Defaults to the `name` input value."
  nullable = true
  default = null
}

variable "virtual_network_name" {
  type = string
  description = "The name of the virtual network. Defaults to the `name` input value."
  nullable = true
  default = null
}

variable "network_security_group_name" {
  type = string
  description = "Specifies the name of the network security group. Defaults to the `name` input value."
  nullable = true
  default = null
}

variable "subnet_name" {
  type = string
  description = "The name of the subnet. Defaults to the `name` input value."
  nullable = true
  default = null
}

variable "network_interface_name" {
  type = string
  description = "The name of the Network Interface. Defaults to the `name` input value."
  nullable = true
  default = null
}

variable "ip_configuration_name" {
  type = string
  description = "A name used for this IP Configuration. Defaults to the `name` input value."
  nullable = true
  default = null
}

variable "virtual_machine_name" {
  type = string
  description = "Virtual machine name. Defaults to the `name` input value."
  nullable = true
  default = null
}

variable "computer_name" {
  type = string
  description = "Specifies the name of the Virtual Machine. Defaults to the `name` input value."
  nullable = true
  default = null
}

variable "storage_os_disk_name" {
  type = string
  description = "Specifies the name of the OS Disk. Defaults to the `name` input value."
  nullable = true
  default = null
}

variable "admin_username" {
  type = string
  description = "Specifies the name of the local administrator account."
  default = "azureuser"
  nullable = false
}

variable "admin_password" {
  type = string
  description = "The password associated with the local administrator account. If empty a random password is generated."
  nullable = true
  default = null
}
