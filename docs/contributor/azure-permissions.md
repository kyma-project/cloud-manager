# Azure Permissions

## Default Principal

```json
{
  "Actions": [
    "Microsoft.Cache/redis/delete",
    "Microsoft.Cache/redis/listKeys/action",
    "Microsoft.Cache/redis/read",
    "Microsoft.Cache/redis/regenerateKey/action",
    "Microsoft.Cache/redis/write",
    "Microsoft.Network/networkSecurityGroups/delete",
    "Microsoft.Network/networkSecurityGroups/read",
    "Microsoft.Network/networkSecurityGroups/write",
    "Microsoft.Network/privateEndpoints/delete",
    "Microsoft.Network/privateEndpoints/read",
    "Microsoft.Network/privateEndpoints/write",
    "Microsoft.Network/virtualNetworks/delete",
    "Microsoft.Network/virtualNetworks/read",
    "Microsoft.Network/virtualNetworks/subnets/delete",
    "Microsoft.Network/virtualNetworks/subnets/read",
    "Microsoft.Network/virtualNetworks/subnets/write",
    "Microsoft.Network/virtualNetworks/write",
    "Microsoft.Resources/subscriptions/resourceGroups/delete",
    "Microsoft.Resources/subscriptions/resourceGroups/read",
    "Microsoft.Resources/subscriptions/resourceGroups/write"
  ]
}
```

## Peering Principal

```json
{
  "Actions": [
    "Microsoft.ClassicNetwork/virtualNetworks/peer/action",
    "Microsoft.Network/virtualNetworks/peer/action",
    "Microsoft.Network/virtualNetworks/read",
    "Microsoft.Network/virtualNetworks/virtualNetworkPeerings/delete",
    "Microsoft.Network/virtualNetworks/virtualNetworkPeerings/read",
    "Microsoft.Network/virtualNetworks/virtualNetworkPeerings/write"
  ]
}
```
