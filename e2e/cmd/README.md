# `e2e` CLI

End-to-end command tool for the Kyma Cloud Manager. The CLI is built from `./e2e/cmd` and is typically invoked as:

```sh
go run ./e2e/cmd [global flags] <command> [<subcommand>] [flags]
```

It drives the same machinery the e2e tests use — provisioning Kyma runtimes via KEB, talking to KCP/SKR clusters, running terraform workspaces against the cloud providers, starting an `envtest` API server, etc.

## Global flags

| Flag | Description |
|------|-------------|
| `--config-dir <path>` | Path to the directory containing `e2econfig.yaml`. Sets the `CONFIG_DIR` env var if provided. Required by most commands that talk to KEB / Garden. |
| `-v`, `--verbose` | Enable verbose logging (zap dev logger). Several commands also enable progress printing when this is set. |
| `-h`, `--help` | Help for any command/subcommand. |

## Command tree

```
e2e
├── cloud
│   └── tf                      run an ad-hoc terraform module against the configured cloud
├── config
│   └── dump                    print the loaded e2e config
├── credentials (c, cred, creds)
│   └── download (down, d)      download configured Garden secrets to --config-dir
├── envtest
│   └── run                     start a local envtest control plane and write a kubeconfig
├── instance
│   ├── clean                   delete SKR cloud-resources from an instance (post-failure cleanup)
│   ├── create                  create a Kyma instance via KEB (optionally wait until ready)
│   ├── credentials (cre, cred, creds)
│   │   ├── dump                print the SKR kubeconfig (and its expiry)
│   │   └── renew               renew the SKR kubeconfig
│   ├── delete                  delete a Kyma instance via KEB
│   ├── ignore (i)              toggle the `do-not-reconcile` label on Runtime/GardenerCluster/Kyma
│   ├── kyma                    print the SKR `Kyma` CR (kyma-system/default)
│   ├── list                    list all KEB instances
│   ├── modules (mo, mod, mods, module)
│   │   ├── add                 add a module to the SKR Kyma spec
│   │   ├── list                list module spec/status across instances
│   │   └── remove              remove a module from the SKR Kyma spec
│   ├── show                    show details for one instance
│   └── wait                    wait for instance(s) to reach completed state
└── sim
    └── run                     run the e2e simulator
```

---

## `cloud tf`

Drives a one-shot terraform workspace: builds it, destroys any previous state, runs `init` → `plan` → `apply`, prints the outputs as JSON, then destroys it again. Uses the configured Garden client and config to construct the cloud workspace.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--source` | `-s` | yes | TF module source, e.g. `terraform-aws-modules/vpc/aws 6.5.1`. |
| `--provider` | `-p` | no, repeatable | TF provider, e.g. `hashicorp/aws ~> 6.0`. |
| `--var` |   | no, repeatable | Input variable in `key="value"` (or `key=123`) form. |
| `--confirm-delete` | `-c` | no | Pause and wait for `[enter]` before the final `destroy`. |

Source: [cloudTf.go](./cloudTf.go)

---

## `config dump`

Prints the loaded `e2econfig.yaml` (after env merging) as YAML to stdout. No flags beyond the globals.

Source: [configDump.go](./configDump.go)

---

## `credentials download` (aliases: `down`, `d`)

Downloads the Garden secrets listed in `config.DownloadGardenSecrets` and writes the requested keys as files under `--config-dir`. Requires `--config-dir` to be set. No command-specific flags.

Source: [credentialsDownload.go](./credentialsDownload.go)

---

## `envtest run`

Starts a local `envtest` Kubernetes API server using the binaries under `<config-dir>/bin/k8s/<version>-<os>-<arch>/`, writes a kubeconfig to `--output`, and then runs until interrupted (Ctrl-C). On exit it stops the env and removes the kubeconfig file.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--output` | `-o` | yes | File path the generated kubeconfig is written to. |
| `--version` |  | yes (or via `ENVTEST_K8S_VERSION`) | Kubernetes version of the envtest assets to use. |

Source: [envtestRun.go](./envtestRun.go)

---

## `instance create`

Creates a new Kyma runtime via KEB with a fresh global/sub account UUID, given alias and provider; optionally waits for provisioning to complete.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--alias` | `-a` | yes | Alias for the instance. |
| `--provider` | `-p` | yes | Provider name (e.g. `aws`, `azure`, `gcp`). |
| `--wait` | `-w` | no | Wait for instance to be ready. |
| `--timeout` | `-t` | no | Wait timeout (default `15m0s`). |

Source: [instanceCreate.go](./instanceCreate.go)

---

## `instance delete`

Marks an instance for deletion via KEB; optionally waits for full destruction. One of `--runtime-id` / `--alias` is required (mutually exclusive).

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--runtime-id` | `-r` | one-of | Runtime ID to delete. |
| `--alias` | `-a` | one-of | Alias of the instance to delete. |
| `--wait` | `-w` | no | Wait for instance to be deleted. |
| `--timeout` | `-t` | no | Wait timeout (default `40m0s`). |

Source: [instanceDelete.go](./instanceDelete.go)

---

## `instance list`

Lists all KEB instances. Output is a table (`default` or `wide`) by default; `json` and `yaml` are also supported.

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | One of `default` (default), `wide`, `json`, `yaml`. The wide format adds Global Account / Sub Account columns. |

Source: [instanceList.go](./instanceList.go)

---

## `instance show`

Prints YAML details for a single instance. Exactly one of `--runtime-id` / `--alias` is required.

| Flag | Short | Description |
|------|-------|-------------|
| `--runtime-id` | `-r` | Runtime ID. |
| `--alias` | `-a` | Alias name. |

Source: [instanceShow.go](./instanceShow.go)

---

## `instance wait`

Waits for instance(s) to reach the completed state. At least one of `--runtime`, `--alias`, `--all` must be given.

| Flag | Short | Description |
|------|-------|-------------|
| `--runtime` | `-r` | Wait for a specific runtime ID. |
| `--alias` | `-a` | Wait for instances matching this alias. |
| `--all` |  | Wait for all runtimes. |
| `--timeout` | `-t` | Wait timeout (default `15m0s`). |

With `--verbose` the wait emits progress lines for each polling iteration.

Source: [instanceWait.go](./instanceWait.go)

---

## `instance clean`

Deletes all SKR cloud-manager resources from a runtime — the cleanup path for orphaned resources after failed e2e tests. By default it deletes everything in the `cloud-resources.kyma-project.io` group **except** the `CloudResources` CR. With `--all` it also wipes the `operator.kyma-project.io` group, including the `Kyma` and `CloudResources` CRs (destructive).

Uses `DeleteAllOf` for efficiency; should run after the test suite but before the instance itself is deleted. One of `--runtime-id` / `--alias` is required (mutually exclusive).

| Flag | Short | Description |
|------|-------|-------------|
| `--runtime-id` | `-r` | Runtime ID of the instance to clean. |
| `--alias` | `-a` | Alias of the instance to clean. |
| `--wait` | `-w` | Wait until everything is deleted. |
| `--timeout` | `-t` | Wait timeout (default `30m0s`). |
| `--force` | `-f` | After timeout, force delete (clears finalizers etc.). |
| `--dry-run` |  | Don't actually delete. |
| `--all` |  | **Destructive.** Also delete the `operator.kyma-project.io` group, i.e. the `Kyma` and `CloudResources` CRs. |

Examples:

```sh
# Clean instance by runtime ID
e2e instance clean --runtime-id abc-123

# Clean instance by alias
e2e instance clean --alias my-test-cluster

# Clean with verbose logging
e2e instance clean --runtime-id abc-123 --verbose
```

Source: [instanceClean.go](./instanceClean.go)

---

## `instance ignore` (alias: `i`)

Toggles the `do-not-reconcile` label on the matching `Runtime`, `GardenerCluster`, and KCP `Kyma` CRs in the control plane (and clears `Kyma.spec.modules` / `Kyma.status.modules` when ignoring). Useful for parking a runtime so reconcilers stop touching it.

Exactly one of the following must be given (mutually exclusive):

| Flag | Short | Description |
|------|-------|-------------|
| `--list-only` | `-l` | List runtimes and their current ignore state — no changes. |
| `--ignore-all` | `-a` | Ignore (disable reconciliation) on **all** runtimes. |
| `--ignore-none` | `-n` | Stop ignoring (enable reconciliation) on **all** runtimes. |
| `--ignore <alias…>` | `-i` | Ignore the listed aliases (or runtime names); enable all others. |
| `--enable <alias…>` | `-e` | Enable the listed aliases; ignore all others. |

Source: [instanceIgnore.go](./instanceIgnore.go)

---

## `instance kyma`

Fetches the SKR `Kyma` CR (`kyma-system/default`) from the runtime and prints it as YAML.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--runtime-id` | `-r` | yes | The runtime ID. |

Source: [instanceKyma.go](./instanceKyma.go)

---

## `instance modules`

Subcommands. Aliases for the group: `mo`, `mod`, `mods`, `module`.

### `instance modules list`

Prints a table of modules for the matching runtime(s) showing whether the module is in `spec`/`status`, its module state, message, and (for `cloud-manager`) the SKR `CloudResources` CR state.

| Flag | Short | Description |
|------|-------|-------------|
| `--runtime-id` | `-r` | Restrict to one runtime. Mutually exclusive with `--alias`. |
| `--alias` | `-a` | Restrict to runtimes matching alias. |
| `--module-name` | `-m` | Filter by module name (repeatable). |

Source: [instanceModulesList.go](./instanceModulesList.go)

### `instance modules add`

Adds a module to the SKR `Kyma` (`kyma-system/default`) `.spec.modules`. If `--wait` is set **and the module is `cloud-manager`**, the command polls until the module reaches the `Ready` state (or `--timeout` elapses).

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--module` | `-m` | yes | Module name to add. |
| `--runtime-id` | `-r` | one-of | Runtime ID. |
| `--alias` | `-a` | one-of | Runtime alias. |
| `--wait` | `-w` | no | Wait until the module is Ready (only honoured for `cloud-manager`). |
| `--timeout` | `-t` | no | Wait timeout (default `5m0s`). |

Source: [instanceModulesAdd.go](./instanceModulesAdd.go)

### `instance modules remove`

Removes a module from the SKR `Kyma` `.spec.modules`. No-op if the module is not present.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--module` | `-m` | yes | Module name to remove. |
| `--runtime-id` | `-r` | one-of | Runtime ID. |
| `--alias` | `-a` | one-of | Runtime alias. |

Source: [instanceModulesRemove.go](./instanceModulesRemove.go)

---

## `instance credentials`

Subcommands. Aliases for the group: `cre`, `cred`, `creds`.

### `instance credentials dump`

Prints the SKR kubeconfig (along with a `# expires at <RFC3339>` comment) to stdout. One of `--runtime-id` / `--alias` is required (mutually exclusive).

| Flag | Short | Description |
|------|-------|-------------|
| `--runtime-id` | `-r` | Runtime ID. |
| `--alias` | `-a` | Runtime alias. |

Source: [instanceCredentialsDump.go](./instanceCredentialsDump.go)

### `instance credentials renew`

Renews the SKR kubeconfig for a runtime via KEB.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--runtime-id` | `-r` | yes | The runtime ID. |

Source: [instanceCredentialsRenew.go](./instanceCredentialsRenew.go)

---

## `sim run`

Runs the e2e simulator (`e2e/sim`) against the loaded config until interrupted or the timeout expires.

| Flag | Short | Description |
|------|-------|-------------|
| `--timeout` | `-t` | Time to run before quitting. `0` (default) means run until Ctrl-C. |

Source: [simRun.go](./simRun.go)
