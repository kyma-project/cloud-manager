
output "private_ip_address" {
  value = azurerm_network_interface.nic.private_ip_address
}

output "vnet_id" {
  // /subscriptions/3f1d2fbd-117a-4742-8bde-6edbcdee6a04/resourceGroups/e2e/providers/Microsoft.Network/virtualNetworks/e2e
  value = "/subscriptions/${data.azurerm_client_config.current.subscription_id}/resourceGroups/${azurerm_resource_group.rg.name}/providers/Microsoft.Network/virtualNetworks/${azurerm_virtual_network.vnet.name}"
}
