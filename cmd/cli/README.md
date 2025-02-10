# CLI Dev Tool

Set the following env vars before running CLI
```shell
export CM_CONTEXT_SKR=kind-skr
export CM_CONTEXT_GARDEN=kind-garden
export CM_CONTEXT_KCP=kind-kcp
```

Create new Shoot in Garden.
```shell
# aws
go run ./cmd/cli garden shoot create aws -p {PROFILE} -s {SHOOT_NAME} -n {NAMESPACE}
# azure
go run ./cmd/cli garden shoot create azure -region {REGION} -u {CLIENT_ID} -p {CLIENT_SECRET} --subscription {SUBSCRIPTION_ID} --tenant {TENANT_ID} -s {SHOOT_NAME} -n {NAMESPACE}
```



Create new Kyma in KCP
```shell
go run ./cmd/cli kyma create -s {SHOOT_NAME} -k {KYMA_NAME}
```

Add Cloud Manager module state to `Ready` in your kyma.
```shell
go run ./cmd/cli kyma module state update -k {KYMA_NAME} -m "cloud-manager" -s Ready
```
