# Dev

# Setup Garden

Generate Gardener core CRDs
```shell
./bin/controller-gen crd:allowDangerousTypes=true \
  paths="$GOPATH/pkg/mod/github.com/gardener/gardener@v1.85.0/pkg/apis/core/v1beta1" \
  output:crd:artifacts:config=tools/dev/gardener
```

Create KCP cluster
```shell
tools/dev/kind/create-kcp.sh
```

Create Garden cluster
```shell
tools/dev/kind/create-garden.sh
```

Note the exported kubeconfig files for both created clusters:
* `tools/dev/kind/kubeconfig-kcp.yaml`
* `tools/dev/kind/kubeconfig-garden.yaml`

Apply the Garden Shoot and SecretBindind CRDs to the Garden cluster
```shell
KUBECONFIG=tools/dev/kind/kubeconfig-garden.yaml \
  kubectl apply \
  -f tools/dev/gardener/core.gardener.cloud_shoots.yaml
KUBECONFIG=tools/dev/kind/kubeconfig-garden.yaml \
  kubectl apply \
  -f tools/dev/gardener/core.gardener.cloud_secretbindings.yaml
```

Create the `gardener-credentials` secret in KCP cluster so you can mount it if running operator in the kind cluster.
```shell
KUBECONFIG=tools/dev/kind/kubeconfig-kcp.yaml \
  kubectl create secret generic \
  -n kcp-system gardener-credentials \
  --from-file=kubeconfig=tools/dev/kind/kubeconfig-garden.yaml
```

Set the env var `GARDENER_CREDENTIALS` to point to the gardener kubeconfig
```shell
export GARDENER_CREDENTIALS=$PROJECT_ROOT/tools/dev/kind/kubeconfig-garden.yaml
```

Set the env var `GARDENER_NAMESPACE` to the value that you will use as garden project and create `Shoot` and `SecretBinding instances` so operator can find them there, ie
```shell
export GARDENER_NAMESPACE=garden-project
```

In the garden cluster create:
* Shoot
* SecretBinding
* Secret

In the kcp cluster create:
* Kyma having the `kyma-project.io/shoot-name` annotation equal to the `Shoot` name you have created in the Garden cluster