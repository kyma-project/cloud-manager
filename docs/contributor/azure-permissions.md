# Azure Permissions

## Default Principal

* Resource Groups
  * `Microsoft.Resources/subscriptions/resourceGroups/read`
  * `Microsoft.Resources/subscriptions/resourceGroups/write`
  * `Microsoft.Resources/subscriptions/resourceGroups/delete`
* VNet
  * `Microsoft.Network/virtualNetworks/read`
  * `Microsoft.Network/virtualNetworks/write`
  * `Microsoft.Network/virtualNetworks/delete`
* Network security group
  * `Microsoft.Network/networkSecurityGroups/read`
  * `Microsoft.Network/networkSecurityGroups/write`
  * `Microsoft.Network/networkSecurityGroups/delete`
* Subnet
  * `Microsoft.Network/virtualNetworks/subnets/read`
  * `Microsoft.Network/virtualNetworks/subnets/write`
  * `Microsoft.Network/virtualNetworks/subnets/delete`
* Private Endpoint
  * `Microsoft.Network/privateEndpoints/read`
  * `Microsoft.Network/privateEndpoints/write`
  * `Microsoft.Network/privateEndpoints/delete`
* Redis
  * `Microsoft.Cache/redis/read`
  * `Microsoft.Cache/redis/write`
  * `Microsoft.Cache/redis/delete`
  * `Microsoft.Cache/redis/listKeys/action`
  * `Microsoft.Cache/redis/regenerateKey/action`

## Peering Principal

* `Microsoft.Network/virtualNetworks/virtualNetworkPeerings/write`
* `Microsoft.Network/virtualNetworks/peer/action`
* `Microsoft.ClassicNetwork/virtualNetworks/peer/action`
* `Microsoft.Network/virtualNetworks/virtualNetworkPeerings/read`
* `Microsoft.Network/virtualNetworks/virtualNetworkPeerings/delete`
* `Microsoft.Network/virtualNetworks/read`
