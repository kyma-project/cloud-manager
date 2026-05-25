// Package managedredis hosts the cross-provider reconciler for the
// KCP AzureManagedRedis CRD.
//
// Despite sharing the directory-naming convention with cross-provider
// packages such as pkg/kcp/redisinstance/ (which has AWS, GCP, and Azure
// providers), this reconciler is intentionally Azure-only — Azure Managed
// Redis is an Azure-specific product (Microsoft.Cache/redisEnterprise).
// The provider switch in reconciler.go has only one case for that reason.
//
// Future contributors: do not add AWS/GCP siblings here. The "managedredis"
// name refers specifically to the Azure Managed Redis service.
package managedredis
