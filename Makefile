.DEFAULT_GOAL := all

.PHONY: all
all:
	@go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.9
	@$(shell go env GOPATH)/bin/controller-gen paths="./..." object crd:crdVersions=v1,trivialVersions=true output:crd:artifacts:config=manifests/crd
