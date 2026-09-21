# Adding Cloud Manager Permissions

This guide explains how to extend Cloud Manager permissions when adding new features that require additional cloud provider permissions.

## Overview

The process is the same for all providers (Azure, GCP, AWS, AliCloud):

1. **Setup** - Authenticate and prepare configuration files
2. **Test** - Run Cloud Manager locally to discover missing permissions
3. **Update** - Add required permissions to provider-specific JSON files
4. **Sync** - Apply permissions to cloud using provider-specific scripts
5. **Request Changes** - Create SRE ticket for production deployment

Only the cloud CLI tools, permission files, and sync scripts differ between providers.

## Common Prerequisites

Before starting, ensure you have:

1. Configuration files:
   - `./tmp/e2e-config.yaml` - Get from `.github/workflows/e2e-tests.yaml` or a colleague
   - `e2e/scripts/.env` - Ask a colleague for a copy

2. Download cloud credentials:
   ```bash
   go run credentials download
   ```
   Credentials will be saved to `./tmp/`

## Azure Permissions

### Prerequisites

- Install Azure CLI: `az`
- Login to SAP shared tenant:
  ```bash
  az login
  ```

### Permission Files

- **Permissions**: `./docs/contributor/permissions/azure_default.json`
- **Conditions**: `./docs/contributor/permissions/azure_default_condition.txt`

### Process

1. **Review current permissions**:
   ```bash
   cat ./docs/contributor/permissions/azure_default.json
   cat ./docs/contributor/permissions/azure_default_condition.txt
   ```

2. **Run Cloud Manager locally**:
   ```bash
   CONFIG_DIR=$(pwd)/tmp go run ./cmd
   ```

3. **Observe 403 permission errors** in logs (e.g., `Microsoft.Network/virtualNetworks/read`)

4. **Add missing permission** to `azure_default.json`:
   ```json
   {
     "permissions": [
       "Microsoft.Network/virtualNetworks/read",
       ...
     ]
   }
   ```

5. **Sync permissions to cloud**:
   ```bash
   ./e2e/scripts/azure_infra.sh
   ```

6. **Restart Cloud Manager** and repeat steps 3-5 until no more 403 errors

7. **Sort permissions** alphabetically:
   ```bash
   ./docs/contributor/permissions/sort.sh
   ```

8. **Create SRE Ticket** to deploy permissions to production:
   - Generate diff from `azure_default.json`
   - Include both the diff and complete final permission set
   - Example: https://github.tools.sap/kyma/backlog/issues/9423

## GCP Permissions

### Prerequisites

- Install Google Cloud SDK: `gcloud`
- Login:
  ```bash
  gcloud auth login
  ```

### Permission Files

- **Permissions**: `./docs/contributor/permissions/gcp_default.json`

### Process

Follow the same process as Azure, but:
- Use `./e2e/scripts/gcp_infra.sh` to sync permissions
- Check `gcp_default.json` for current permissions

## AWS Permissions

### Prerequisites

- Install AWS CLI: `aws`
- Install `aws-azure-login` (see `tools/dev/aws/README.md`)
- Configure AWS profile for SAP subscription

### Permission Files

- **Permissions**: `./docs/contributor/permissions/aws_default.json`

### Process

Follow the same process as Azure, but:
- Use `./e2e/scripts/aws_infra.sh` to sync permissions
- Check `aws_default.json` for current permissions

## Quick Reference

| Provider | CLI Tools | Permission File | Sync Script |
|----------|-----------|----------------|-------------|
| Azure | `az` | `azure_default.json` | `azure_infra.sh` |
| GCP | `gcloud` | `gcp_default.json` | `gcp_infra.sh` |
| AWS | `aws`, `aws-azure-login` | `aws_default.json` | `aws_infra.sh` |
| AliCloud | `aliyun` | `alicloud/policy-CloudManagerAccess.json` | `alicloud_infra.sh` |

## Tips

- Always run `./docs/contributor/permissions/sort.sh` before committing changes
- Keep the SRE ticket updated with complete permission sets, not just diffs
- Test thoroughly - iterate until all 403 errors are resolved
- Document why new permissions are needed in the SRE ticket
