
# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.33.0
JV_VERSION = v0.5.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	./config/patchAfterMakeManifests.sh

.PHONY: mod-download
mod-download:
	go mod download

.PHONY: garden-manifests
garden-manifests: mod-download
	rm -r config/crd/gardener-core-tmp || true
	$(CONTROLLER_GEN) crd:allowDangerousTypes=true paths="$(shell go list -m -f '{{.Dir}}'  github.com/gardener/gardener)/pkg/apis/core/v1beta1/..." output:crd:artifacts:config=config/crd/gardener-core-tmp
	cp config/crd/gardener-core-tmp/core.gardener.cloud_shoots.yaml config/crd/gardener/core.gardener.cloud_shoots.yaml
	cp config/crd/gardener-core-tmp/core.gardener.cloud_secretbindings.yaml config/crd/gardener/core.gardener.cloud_secretbindings.yaml
	cp config/crd/gardener-core-tmp/core.gardener.cloud_cloudprofiles.yaml config/crd/gardener/core.gardener.cloud_cloudprofiles.yaml
	rm -r config/crd/gardener-core-tmp
	rm -r config/crd/gardener-security-tmp || true
	$(CONTROLLER_GEN) crd:allowDangerousTypes=true paths="$(shell go list -m -f '{{.Dir}}'  github.com/gardener/gardener)/pkg/apis/security/v1alpha1/..." output:crd:artifacts:config=config/crd/gardener-security-tmp
	cp config/crd/gardener-security-tmp/security.gardener.cloud_credentialsbindings.yaml config/crd/gardener/security.gardener.cloud_credentialsbindings.yaml
	rm -r config/crd/gardener-security-tmp

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test-ff
test-ff: jv download-flag-schema
	$(LOCALBIN)/jv -assertcontent -assertformat "$(GO_FF_SCHEMA_FILE)" ./pkg/feature/ff_ga.yaml
	$(LOCALBIN)/jv -assertcontent -assertformat "$(GO_FF_SCHEMA_FILE)" ./pkg/feature/ff_edge.yaml

.PHONY: test
test: manifests generate fmt vet envtest test-ff build_ui ## Run tests.
	SKR_PROVIDERS="$(PROJECTROOT)/config/dist/skr/bases/providers" ENVTEST_K8S_VERSION="$(ENVTEST_K8S_VERSION)" PROJECTROOT="$(PROJECTROOT)" KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" GOFIPS140=v1.0.0 go test ./... -test.v -v -coverprofile cover.out

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.54.2
golangci-lint:
	@[ -f $(GOLANGCI_LINT) ] || { \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell dirname $(GOLANGCI_LINT)) $(GOLANGCI_LINT_VERSION) ;\
	}

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

##@ Build

.PHONY: build
build: manifests generate fmt vet build_ui ## Build manager binary.
	GOFIPS140=v1.0.0 go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	GODEBUG=fips140=only,tlsmlkem=0 go run ./cmd/main.go

# If you wish to build the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

# PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - be able to use docker buildx. More info: https://docs.docker.com/build/buildx/
# - have enabled BuildKit. More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image to your registry (i.e. if you do not set a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To adequately provide solutions that are compatible with multiple platforms, you should consider using this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name project-v3-builder
	$(CONTAINER_TOOL) buildx use project-v3-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm project-v3-builder
	rm Dockerfile.cross

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/dist/kcp/crd | $(KUBECTL) apply -f -

.PHONY: build_ui
build_ui: manifests kustomize # Build CRDS test
	# kustomize build all the ConfigMaps and output to their own file
	@$(KUSTOMIZE) build config/ui-extensions/gcpnfsvolumes > config/ui-extensions/gcpnfsvolumes/cloud-resources.kyma-project.io_gcpnfsvolumes_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/gcpnfsvolumebackups > config/ui-extensions/gcpnfsvolumebackups/cloud-resources.kyma-project.io_gcpnfsvolumebackups_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/gcpnfsvolumerestores > config/ui-extensions/gcpnfsvolumerestores/cloud-resources.kyma-project.io_gcpnfsvolumerestores_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/ipranges > config/ui-extensions/ipranges/cloud-resources.kyma-project.io_ipranges_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/gcpvpcpeerings > config/ui-extensions/gcpvpcpeerings/cloud-resources.kyma-project.io_gcpvpcpeerings_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/gcpredisinstances > config/ui-extensions/gcpredisinstances/cloud-resources.kyma-project.io_gcpredisinstances_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/gcpnfsbackupschedules > config/ui-extensions/gcpnfsbackupschedules/cloud-resources.kyma-project.io_gcpnfsbackupschedules_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/gcpredisclusters > config/ui-extensions/gcpredisclusters/cloud-resources.kyma-project.io_gcpredisclusters_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/gcpsubnets > config/ui-extensions/gcpsubnets/cloud-resources.kyma-project.io_gcpsubnets_ui.yaml

	@$(KUSTOMIZE) build config/ui-extensions/awsnfsvolumes > config/ui-extensions/awsnfsvolumes/cloud-resources.kyma-project.io_awsnfsvolumes_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/awsredisinstances > config/ui-extensions/awsredisinstances/cloud-resources.kyma-project.io_awsredisinstances_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/awsvpcpeerings > config/ui-extensions/awsvpcpeerings/cloud-resources.kyma-project.io_awsvpcpeerings_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/awsnfsvolumebackups > config/ui-extensions/awsnfsvolumebackups/cloud-resources.kyma-project.io_awsnfsvolumebackups_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/awsnfsvolumerestores > config/ui-extensions/awsnfsvolumerestores/cloud-resources.kyma-project.io_awsnfsvolumerestores_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/awsnfsbackupschedules > config/ui-extensions/awsnfsbackupschedules/cloud-resources.kyma-project.io_awsnfsbackupschedules_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/awsredisclusters > config/ui-extensions/awsredisclusters/cloud-resources.kyma-project.io_awsredisclusters_ui.yaml

	@$(KUSTOMIZE) build config/ui-extensions/azurevpcpeerings > config/ui-extensions/azurevpcpeerings/cloud-resources.kyma-project.io_azurevpcpeerings_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/azureredisinstances > config/ui-extensions/azureredisinstances/cloud-resources.kyma-project.io_azureredisinstances_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/azurerwxbackupschedules > config/ui-extensions/azurerwxbackupschedules/cloud-resources.kyma-project.io_azurerwxbackupschedules_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/azurerwxvolumerestores > config/ui-extensions/azurerwxvolumerestores/cloud-resources.kyma-project.io_azurerwxvolumerestores_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/azureredisclusters > config/ui-extensions/azureredisclusters/cloud-resources.kyma-project.io_azureredisclusters_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/azurevpcdnslinks > config/ui-extensions/azurevpcdnslinks/cloud-resources.kyma-project.io_azurevpcdnslinks_ui.yaml
	@$(KUSTOMIZE) build config/ui-extensions/sapnfsvolumes > config/ui-extensions/sapnfsvolumes/cloud-resources.kyma-project.io_sapnfsvolumes_ui.yaml



.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/dist/kcp/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

##@ Build Dependencies

PROJECTROOT = $(shell pwd)

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

GO_FF_VERSION := $(shell grep "github.com/thomaspoignant/go-feature-flag " ./go.mod | awk '{print $$2}')
GO_FF_SCHEMA_URL := https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/$(GO_FF_VERSION)/.schema/flag-schema.json
GO_FF_SCHEMA_FILE := $(LOCALBIN)/flag-schema-$(GO_FF_VERSION).json

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
JV ?= $(LOCALBIN)/jv

## Tool Versions
KUSTOMIZE_VERSION ?= v5.5.0
CONTROLLER_TOOLS_VERSION ?= v0.16.5

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || GOBIN=$(LOCALBIN) GO111MODULE=on go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: jv
jv:$(JV)
$(JV): $(LOCALBIN)
	test -s $(LOCALBIN)/jv  || \
	GOBIN=$(LOCALBIN) go install github.com/santhosh-tekuri/jsonschema/cmd/jv@$(JV_VERSION)

.PHONY: download-flag-schema
download-flag-schema: $(GO_FF_SCHEMA_FILE)

$(GO_FF_SCHEMA_FILE):
	curl -sSL -o  $(GO_FF_SCHEMA_FILE) "$(GO_FF_SCHEMA_URL)"
