# CLI Dev Tool

Set the following env vars before running CLI
```shell
export CM_CONTEXT_SKR=kind-skr
export CM_CONTEXT_GARDEN=kind-garden
export CM_CONTEXT_KCP=kind-kcp
```

Create new Shoot in Garden.
```shell
go run ./cmd/cli garden shoot create aws -p skr -s my-shoot
```

Create new Kyma in KCP
```shell
go run ./cmd/cli kyma create -s my-shoot -k my-kyma.
```

Enable Cloud Manager module state to Ready in your kyma.
```shell
go run ./cmd/cli kyma module state update -k my-kyma -m "cloud-manager" -s Ready
```
