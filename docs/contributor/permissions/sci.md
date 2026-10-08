# SCI Permissions

SCI permissions are not managed via a JSON file and sync script like the other providers. They live on the CCEE technical users, documented in kyma/backlog issue 6704.

## TL;DR

Last info: Cloud Manager reuses the existing Gardener CCEE technical users (`TKYMA_{DEV,STG,PRD}_{001,002}`), with only the `sharedfilesystem_admin` (Manila Administrator) role added so it can manage OpenStack SharedFileSystem shares.

For full context (roles, regions, projects, Vault changes) read kyma/backlog issue 6704.
