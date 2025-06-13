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
    "Microsoft.Network/natGateways/read",
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
    "Microsoft.Network/publicIPAddresses/read",
    "Microsoft.Network/virtualNetworks/delete",
    "Microsoft.Network/virtualNetworks/join/action",
    "Microsoft.Network/virtualNetworks/read",
    "Microsoft.Network/virtualNetworks/subnets/delete",
    "Microsoft.Network/virtualNetworks/subnets/join/action",
    "Microsoft.Network/virtualNetworks/subnets/read",
    "Microsoft.Network/virtualNetworks/subnets/write",
    "Microsoft.Network/virtualNetworks/write",
    "Microsoft.RecoveryServices/Vaults/backupFabrics/protectionContainers/delete",
    "Microsoft.RecoveryServices/Vaults/backupFabrics/protectionContainers/protectedItems/delete",
    "Microsoft.RecoveryServices/Vaults/backupFabrics/protectionContainers/protectedItems/operationResults/read",
    "Microsoft.RecoveryServices/Vaults/backupFabrics/protectionContainers/protectedItems/operationsStatus/read",
    "Microsoft.RecoveryServices/Vaults/backupFabrics/protectionContainers/protectedItems/read",
    "Microsoft.RecoveryServices/Vaults/backupFabrics/protectionContainers/protectedItems/recoveryPoints/restore/action",
    "Microsoft.RecoveryServices/Vaults/backupFabrics/protectionContainers/protectedItems/write",
    "Microsoft.RecoveryServices/Vaults/backupFabrics/protectionContainers/read",
    "Microsoft.RecoveryServices/Vaults/backupFabrics/protectionContainers/write",
    "Microsoft.RecoveryServices/Vaults/backupJobs/operationResults/read",
    "Microsoft.RecoveryServices/Vaults/backupJobs/read",
    "Microsoft.RecoveryServices/Vaults/backupPolicies/read",
    "Microsoft.RecoveryServices/Vaults/backupPolicies/write",
    "Microsoft.RecoveryServices/Vaults/backupProtectableItems/read",
    "Microsoft.RecoveryServices/Vaults/backupProtectedItems/read",
    "Microsoft.RecoveryServices/Vaults/backupconfig/read",
    "Microsoft.RecoveryServices/Vaults/backupconfig/write",
    "Microsoft.RecoveryServices/Vaults/delete",
    "Microsoft.RecoveryServices/Vaults/read",
    "Microsoft.RecoveryServices/Vaults/registeredIdentities/operationResults/read",
    "Microsoft.RecoveryServices/Vaults/registeredIdentities/read",
    "Microsoft.RecoveryServices/Vaults/registeredIdentities/write",
    "Microsoft.RecoveryServices/Vaults/write",
    "Microsoft.Resources/subscriptions/resourceGroups/delete",
    "Microsoft.Resources/subscriptions/resourceGroups/read",
    "Microsoft.Resources/subscriptions/resourceGroups/write",
    "Microsoft.Storage/storageAccounts/fileServices/shares/delete",
    "Microsoft.Storage/storageAccounts/fileServices/shares/read",
    "Microsoft.Storage/storageAccounts/fileServices/shares/write",
    "Microsoft.Storage/storageAccounts/listKeys/action",
    "Microsoft.Storage/storageAccounts/read",
    "Microsoft.Storage/storageAccounts/write"
  ],
  "dataActions": [
    "Microsoft.Storage/storageAccounts/fileServices/fileshares/files/delete",
    "Microsoft.Storage/storageAccounts/fileServices/fileshares/files/write",
    "Microsoft.Storage/storageAccounts/fileServices/fileshares/files/read"
  ]
}
```

## Peering Principal

```json
{
  "Actions": [
    "Microsoft.ClassicNetwork/virtualNetworks/peer/action",
    "Microsoft.Network/virtualNetworks/join/action",
    "Microsoft.Network/virtualNetworks/peer/action",
    "Microsoft.Network/virtualNetworks/read",
    "Microsoft.Network/virtualNetworks/virtualNetworkPeerings/delete",
    "Microsoft.Network/virtualNetworks/virtualNetworkPeerings/read",
    "Microsoft.Network/virtualNetworks/virtualNetworkPeerings/write"
  ]
}
```
