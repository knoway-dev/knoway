GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
export GOPROXY ?= http://10.6.100.13:8081/repository/go-proxy/
export GOSUMDB ?= sum.golang.org http://10.6.100.13:8081/repository/gosum/

HUB ?= release-ci.daocloud.io
HUB_CI = release-ci.daocloud.io

HELM_REPO ?= https://release-ci.daocloud.io/chartrepo/knoway

PROD_NAME = knoway

MINOR_VERSION ?= 0.1
VERSION ?= $(MINOR_VERSION).0-dev.$(shell git rev-parse --short=8 HEAD)

REGISTRY_USER_NAME ?=
REGISTRY_PASSWORD ?=

PUSH_IMAGES ?= 1
OFFLINE ?= 0

PLATFORMS ?= linux/amd64,linux/arm64

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen-$(CONTROLLER_TOOLS_VERSION)
CONTROLLER_TOOLS_VERSION ?= v0.16.5

GOLANGCI_LINT_VERSION ?= v1.57.2
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

ifeq ($(shell uname),Darwin)
SEDI=sed -i ""
else
SEDI=sed -i
endif

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: help ## Display this help.
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} \
	/^[.][Pp][Hh][Oo][Nn][Yy]:[[:space:]]*[a-zA-Z_0-9-]+[[:space:]]*##/ { \
	    split($$0, parts, "##"); \
	    sub("[.][Pp][Hh][Oo][Nn][Yy]:[[:space:]]*", "", parts[1]); \
	    command = parts[1]; \
	    comment = parts[2]; \
	    gsub(/^[[:space:]]+|[[:space:]]+$$/, "", command); \
	    gsub(/^[[:space:]]+|[[:space:]]+$$/, "", comment); \
	    printf "  \033[36m%-15s\033[0m %s\n", command, comment; \
	} \
	/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: format ## Run format Go and Proto
format: format-proto format-go

format-go:
	goimports -local knoway.dev -w .
	gofmt -w .

format-proto:
	cd api; find . -name "*.proto" -exec clang-format -style=file -i {} \;

.PHONY: lint ## Run golangci-lint linter
lint: golangci-lint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix ## Run golangci-lint linter and perform fixes
lint-fix: golangci-lint
	$(GOLANGCI_LINT) run --fix

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,${GOLANGCI_LINT_VERSION})

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef

gen-proto:
	cd api; buf generate --timeout 10m -v \
		--path filters \
		--path listeners \
		--path clusters \
		--path route \
		--path service

gen: gen-proto gen-crds format
.PHONY: gen

.PHONY: clean-proto
clean-proto:
	cd api; ./clean.sh


.PHONY: gen-crds
gen-crds: manifests generate format

.PHONY: download-deps
# Download dependencies with specified versions
download-deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0

.PHONY: controller-gen ## Download controller-gen locally if necessary.
controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))


.PHONY: manifests ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd:ignoreUnexportedFields=true \
 	webhook paths="./..." output:crd:artifacts:config=config/crd/bases; \
 	bash ./scripts/copy-crds.sh config/crd/bases/llm.knoway.dev_llmbackends.yaml manifests/knoway/templates


.PHONY: generate ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

pre-docker-buildx:
	@echo "build and push images to $(HUB)"
ifneq ($(REGISTRY_USER_NAME),)
	docker login -u $${REGISTRY_USER_NAME} -p $${REGISTRY_PASSWORD} ${HUB}
endif
	docker buildx create --name knoway-builder --use || true


APP = knoway-gateway

.PHONY: build-binaries ## Build or download manager binaries.
build-binaries: manifests generate format
	APP=$(APP) PLATFORMS=$(PLATFORMS) bash ./scripts/build-or-download-binaries.sh

ifeq ($(PUSH_IMAGES),1)
BUILD_CMD=buildx build --platform $(PLATFORMS) --push
else
BUILD_CMD=buildx build --platform $(PLATFORMS)
endif

images: build-binaries
	docker $(BUILD_CMD) $(DOCKER_BUILD_FLAGS) \
		--build-arg HUB=$(HUB) \
		--build-arg VERSION=$(VERSION) \
		--build-arg APP=$(APP) \
		-t $(HUB)/$(PROD_NAME)/$(APP):$(VERSION) \
		-f Dockerfile .

.PHONY: release ## Build and push Docker images and Helm charts
release: images push-chart

define in_place_replace
	yq eval $(1) $(2) -i
endef


helm:
	@rm -rf dist/$(PROD_NAME) && mkdir -p dist/$(PROD_NAME)
	@cp -rf manifests/knoway/. dist/$(PROD_NAME)
	$(SEDI) 's/version: .*/version: $(VERSION) # auto generated from build version/g' dist/$(PROD_NAME)/Chart.yaml
	@if [ $(OFFLINE) = 1 ]; then \
        $(SEDI) 's/release.daocloud.io/release-ci.daocloud.io/g' dist/$(PROD_NAME)/values.yaml; \
    fi
	$(call in_place_replace, '.gateway.image.tag = "$(VERSION)"', dist/$(PROD_NAME)/values.yaml)
	$(call in_place_replace, '.global.imageRegistry = "$(HUB)"', dist/$(PROD_NAME)/values.yaml)

	helm package dist/$(PROD_NAME) -d dist --version=$(VERSION)
	@#rm -rf dist/$(PROD_NAME)

.PHONY: push-chart
push-chart: helm
	helm repo add knoway-release $(HELM_REPO)
	helm cm-push ./dist/$(PROD_NAME)-$(VERSION).tgz knoway-release -a $(VERSION) -v $(VERSION) -u $${REGISTRY_USER_NAME} -p $${REGISTRY_PASSWORD} --timeout 300

	# push oci
	helm registry login $(HUB) -u $${REGISTRY_USER_NAME} -p $${REGISTRY_PASSWORD}
	helm push ./dist/$(PROD_NAME)-$(VERSION).tgz oci://$(HUB)/$(PROD_NAME)

.PHONY: unit-test
unit-test:
	bash ./scripts/unit-test.sh

.PHONY: helm-render-check
helm-render-check:
	@for c in manifests/*; do \
		helm template $$c > /dev/null || exit 1; \
	done

.PHONY: license-lint
license-lint:
	scripts/verify-license.sh

.PHONY: staticcheck
staticcheck: license-lint helm-render-check
	scripts/verify-staticcheck.sh

.PHONY: gen-check
gen-check:
	scripts/gen-check.sh

.PHONY: security-scanning
security-scanning:
	bash ./scripts/trivy.sh $(VULNEEABILITY_LEVEL) \
	$(HUB)/$(PROD_NAME)/$(APP):$(VERSION)

