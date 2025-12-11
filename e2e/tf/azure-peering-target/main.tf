
data "azurerm_client_config" "current" {}

resource "random_string" "admin_password" {
  length = 16
  special = true
}

locals {
  resource_group_name = coalesce(tostring(var.resource_group_name), var.name)
  virtual_network_name = coalesce(tostring(var.virtual_network_name), var.name)
  network_security_group_name = coalesce(tostring(var.network_security_group_name), var.name)
  subnet_name = coalesce(tostring(var.subnet_name), var.name)
  network_interface_name = coalesce(tostring(var.network_interface_name), var.name)
  ip_configuration_name = coalesce(tostring(var.ip_configuration_name), var.name)
  virtual_machine_name = coalesce(tostring(var.virtual_machine_name), var.name)
  computer_name = coalesce(tostring(var.computer_name), var.name)
  storage_os_disk_name = coalesce(tostring(var.storage_os_disk_name), var.name)
  admin_password = coalesce(tostring(var.admin_password), random_string.admin_password.result)

  subnet_address_prefix = coalesce(tostring(var.subnet_address_prefix), var.virtual_network_address_space)
}

resource "azurerm_resource_group" "rg" {
  name     = local.resource_group_name
  location = var.location
}

resource "azurerm_virtual_network" "vnet" {
  name                = local.virtual_network_name
  address_space       = [var.virtual_network_address_space]
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  tags = {
    "e2e":"e2e"
  }
}

resource "azurerm_network_security_group" "nsg" {
  name = local.network_security_group_name
  location = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
}

resource "azurerm_subnet" "subnet" {
  name                 = local.subnet_name
  address_prefixes     = [local.subnet_address_prefix]
  virtual_network_name = azurerm_virtual_network.vnet.name
  resource_group_name  = azurerm_resource_group.rg.name
}

resource "azurerm_subnet_network_security_group_association" "nsga" {
  subnet_id = azurerm_subnet.subnet.id
  network_security_group_id = azurerm_network_security_group.nsg.id
}

resource "azurerm_network_interface" "nic" {
  name = local.network_interface_name
  resource_group_name = azurerm_resource_group.rg.name
  location = azurerm_resource_group.rg.location
  ip_configuration {
    name = local.ip_configuration_name
    private_ip_address_allocation = "Dynamic"
    subnet_id = azurerm_subnet.subnet.id
  }
}

resource "azurerm_virtual_machine" "vm" {
  name = local.virtual_machine_name
  resource_group_name = azurerm_resource_group.rg.name
  location =  azurerm_resource_group.rg.location
  vm_size = "Standard_B1s"
  network_interface_ids = [ azurerm_network_interface.nic.id ]
  delete_os_disk_on_termination = true
  delete_data_disks_on_termination = true
  storage_image_reference {
    publisher = "Canonical"
    offer = "0001-com-ubuntu-server-jammy"
    sku = "22_04-lts"
    version = "latest"
  }
  storage_os_disk {
    name = local.storage_os_disk_name
    create_option = "FromImage"
    managed_disk_type = "Standard_LRS"
    caching = "ReadWrite"
  }
  os_profile {
    computer_name = local.computer_name
    admin_username = var.admin_username
    admin_password = local.admin_password
    custom_data = file("${path.module}/cloud-init-docker.txt")
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}
