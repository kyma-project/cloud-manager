# Azure Permissions

## Default Principal

```json
{
  "Actions": [
    "Microsoft.Cache/redis/delete",
    "Microsoft.Cache/redis/listKeys/action",
    "Microsoft.Cache/redis/privateEndpointConnectionsApproval/action",
    "Microsoft.Cache/redis/read",
    "Microsoft.Cache/redis/regenerateKey/action",
    "Microsoft.Cache/redis/write",
    "Microsoft.DataProtection/backupVaults/backupInstances/backup/action",
    "Microsoft.DataProtection/backupVaults/backupInstances/restore/action",
    "Microsoft.Network/networkSecurityGroups/delete",
    "Microsoft.Network/networkSecurityGroups/join/action",
    "Microsoft.Network/networkSecurityGroups/read",
    "Microsoft.Network/networkSecurityGroups/write",
    "Microsoft.Network/privateDnsZones/delete",
    "Microsoft.Network/privateDnsZones/join/action",
    "Microsoft.Network/privateDnsZones/read",
    "Microsoft.Network/privateDnsZones/virtualNetworkLinks/delete",
    "Microsoft.Network/privateDnsZones/virtualNetworkLinks/read",
    "Microsoft.Network/privateDnsZones/virtualNetworkLinks/write",
    "Microsoft.Network/privateDnsZones/write",
    "Microsoft.Network/privateEndpoints/delete",
    "Microsoft.Network/privateEndpoints/privateDnsZoneGroups/delete",
    "Microsoft.Network/privateEndpoints/privateDnsZoneGroups/read",
    "Microsoft.Network/privateEndpoints/privateDnsZoneGroups/write",
    "Microsoft.Network/privateEndpoints/read",
    "Microsoft.Network/privateEndpoints/write",
    "Microsoft.Network/virtualNetworks/delete",
    "Microsoft.Network/virtualNetworks/join/action",
    "Microsoft.Network/virtualNetworks/read",
    "Microsoft.Network/virtualNetworks/subnets/delete",
    "Microsoft.Network/virtualNetworks/subnets/join/action",
    "Microsoft.Network/virtualNetworks/subnets/read",
    "Microsoft.Network/virtualNetworks/subnets/write",
    "Microsoft.Network/virtualNetworks/write",
    "Microsoft.RecoveryServices/Vaults/backupFabrics/protectionContainers/protectedItems/read",
    "Microsoft.RecoveryServices/Vaults/backupFabrics/protectionContainers/protectedItems/write",
    "Microsoft.RecoveryServices/Vaults/backupJobs/operationResults/read",
    "Microsoft.RecoveryServices/Vaults/backupJobs/read",
    "Microsoft.RecoveryServices/Vaults/backupPolicies/read",
    "Microsoft.RecoveryServices/Vaults/backupPolicies/write",
    "Microsoft.RecoveryServices/Vaults/backupProtectableItems/read",
    "Microsoft.RecoveryServices/Vaults/backupProtectedItems/read",
    "Microsoft.RecoveryServices/Vaults/read",
    "Microsoft.RecoveryServices/Vaults/write",
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
