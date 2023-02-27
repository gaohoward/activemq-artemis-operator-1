# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 7.11.0-opr-1

KUBE_CLI=kubectl
OPERATOR_VERSION := 7.11-8
OPERATOR_ACCOUNT_NAME := amq-broker-operator
OPERATOR_CLUSTER_ROLE_NAME := operator-role
OPERATOR_IMAGE_REPO := registry.redhat.io/amq7/amq-broker-rhel8-operator
OPERATOR_NAMESPACE := amq-broker-operator
BUNDLE_PACKAGE := $(OPERATOR_NAMESPACE)
BUNDLE_ANNOTATION_PACKAGE := amq-broker-rhel8
GO_MODULE := github.com/artemiscloud/activemq-artemis-operator
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)


# directory to hold static resources for deploying operator
DEPLOY := ./deploy

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
CHANNELS = "7.11.x"
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
DEFAULT_CHANNEL = "7.11.x"
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# amq.io/e-sdk15-bundle:$VERSION and amq.io/e-sdk15-catalog:$VERSION.
IMAGE_TAG_BASE ?= quay.io/artemiscloud/activemq-artemis-operator

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# Image URL to use all building/pushing image targets
IMG ?= $(OPERATOR_IMAGE_REPO):$(OPERATOR_VERSION)
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.22

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

BUILD_TIMESTAMP := $(shell date '+%Y-%m-%dT%H:%M:%S')
LDFLAGS = "'$(GO_MODULE)/version.BuildTimestamp=$(BUILD_TIMESTAMP)'"


all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen kustomize
ifeq ($(ENABLE_WEBHOOKS),true)
## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
## v2alpha3, v2alpha4 and v2alpha3 requires allowDangerousTypes=true because they use float32 type
	cd config/manager && $(KUSTOMIZE) edit add resource webhook_secret.yaml 
	$(CONTROLLER_GEN) rbac:roleName=$(OPERATOR_CLUSTER_ROLE_NAME) crd:allowDangerousTypes=true webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	find config -type f -exec sed -i -e '/creationTimestamp/d' {} \;
else
## Generate ClusterRole and CustomResourceDefinition objects.
## v2alpha3, v2alpha4 and v2alpha3 requires allowDangerousTypes=true because they use float32 type
	cd config/manager && $(KUSTOMIZE) edit remove resource webhook_secret.yaml 
	$(CONTROLLER_GEN) rbac:roleName=$(OPERATOR_CLUSTER_ROLE_NAME) crd:allowDangerousTypes=true paths="./..." output:crd:artifacts:config=config/crd/bases
	find config -type f -exec sed -i -e '/creationTimestamp/d' {} \;
endif

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet -composites=false ./...

## Run tests.
test test-v: TEST_VARS = KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" RECONCILE_RESYNC_PERIOD=5s

## Run tests against minikube with local operator.
test-mk test-mk-v: TEST_ARGS += -test.timeout=50m -ginkgo.label-filter='!do'
test-mk test-mk-v: TEST_VARS = ENABLE_WEBHOOKS=false USE_EXISTING_CLUSTER=true RECONCILE_RESYNC_PERIOD=5s

## Run tests against minikube with deployed operator(do)
test-mk-do test-mk-do-v: TEST_ARGS += -test.timeout=40m -ginkgo.label-filter='do'
test-mk-do test-mk-do-v: TEST_VARS = DEPLOY_OPERATOR=true ENABLE_WEBHOOKS=false USE_EXISTING_CLUSTER=true

## Run tests against minikube with deployed operator(do) and exclude slow, useful for CI smoke
test-mk-do-fast test-mk-do-fast-v: TEST_ARGS += -test.timeout=40m -ginkgo.label-filter='do && !slow'
test-mk-do-fast test-mk-do-fast-v: TEST_VARS = DEPLOY_OPERATOR=true ENABLE_WEBHOOKS=false USE_EXISTING_CLUSTER=true

test-v test-mk-v test-mk-do-v test-mk-do-fast-v: TEST_ARGS += -v
test-v test-mk test-mk-v test-mk-do test-mk-do-v test-mk-do-fast test-mk-do-fast-v: TEST_ARGS += -ginkgo.slow-spec-threshold=30s -ginkgo.fail-fast -coverprofile cover-mk.out

test test-v test-mk test-mk-v test-mk-do test-mk-do-v test-mk-do-fast test-mk-do-fast-v: manifests generate fmt vet envtest 
	$(TEST_VARS) go test ./... -p 1 $(TEST_ARGS)

##@ Build

build: generate fmt vet manifests ## Build manager binary.
	go build -ldflags=$(LDFLAGS) -o bin/manager main.go

run: manifests generate fmt vet ## Run a controller from your host.
	go run -ldflags=$(LDFLAGS) ./main.go

docker-build: test generate-deploy ## Build docker image with the manager.
	docker build -t ${IMG} .

docker-push: ## Push docker image with the manager.
	docker push ${IMG}

podman-remote-build: build generate-deploy
	podman-remote build -t ${IMG} .

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBE_CLI) apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBE_CLI) delete -f -

deploy: manifests kustomize generate-deploy ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBE_CLI) apply -f -

deploy-dry-run: manifests kustomize ## Create deploy yaml file in tmp/deploy.yaml
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	mkdir -p tmp && $(KUSTOMIZE) build config/default > tmp/deploy.yaml

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | $(KUBE_CLI) delete -f -

generate-deploy: manifests kustomize ## Generate deployment artifacts in separate files in $(DEPLOY) dir
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | hack/static_manifest_gen.sh $(DEPLOY) $(OPERATOR_NAMESPACE)


## Download tools locally if necessary.
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen:
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize:
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

ENVTEST = $(shell pwd)/bin/setup-envtest
envtest:
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: bundle
bundle: manifests operator-sdk kustomize ## Generate bundle manifests and metadata, then validate generated files.
	$(OPERATOR_SDK) generate kustomize manifests -q --package $(BUNDLE_PACKAGE)
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --package $(BUNDLE_PACKAGE) --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	sed -i '/creationTimestamp/d' ./bundle/manifests/*.yaml
	sed 's/annotations://' config/metadata/$(BUNDLE_PACKAGE).annotations.yaml >> bundle/metadata/annotations.yaml
	sed -e 's/annotations://' -e 's/  /LABEL /g' -e 's/: /=/g'  config/metadata/$(BUNDLE_PACKAGE).annotations.yaml >> bundle.Dockerfile
	sed -i 's/operators.operatorframework.io.bundle.package.v1:.*/operators.operatorframework.io.bundle.package.v1: $(BUNDLE_ANNOTATION_PACKAGE)/' bundle/metadata/annotations.yaml
	sed -i 's/operators.operatorframework.io.bundle.package.v1=.*/operators.operatorframework.io.bundle.package.v1=$(BUNDLE_ANNOTATION_PACKAGE)/' bundle.Dockerfile
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: bundle-clean
bundle-clean: ## Clean the bundle directory
	rm -rf ./bundle/*

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.1/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

.PHONY: operator-sdk
OPERATOR_SDK = $(shell pwd)/bin/operator-sdk
operator-sdk: ## Download operator-sdk locally if necessary.
ifeq (,$(wildcard $(OPERATOR_SDK)))
ifeq (,$(shell which operator-sdk 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPERATOR_SDK)) ;\
	curl -sSLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/v1.15.0/operator-sdk_${OS}_${ARCH} ;\
	chmod +x $(OPERATOR_SDK) ;\
	}
else
OPERATOR_SDK = $(shell which operator-sdk)
endif	
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)
