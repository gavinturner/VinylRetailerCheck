ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: generate
generate:
	# TOP TIP: "make generate pkg=[package name]" runs go generate for a specific package
	@go mod vendor
	@docker build -t makegenerate_vinylretailers - < Dockerfile.generate
ifdef pkg
	@docker run --rm -it -e GO111MODULE=off -e GOPATH=$(HOME)/go -v $(HOME)/go:$(HOME)/go:delegated -v $(ROOT_DIR):$(HOME)/go/src/github.com/gavinturner/vinylretailers:delegated -w $(HOME)/go/src/github.com/gavinturner/vinylretailers makegenerate_vinylretailers go generate -mod vendor -x ./$(pkg)/...
else
	@docker run --rm -it -e GO111MODULE=off -e GOPATH=$(HOME)/go -v $(HOME)/go:$(HOME)/go:delegated -v $(ROOT_DIR):$(HOME)/go/src/github.com/gavinturner/vinylretailers:delegated -w $(HOME)/go/src/github.com/gavinturner/vinylretailers makegenerate_vinylretailers go generate -mod vendor -x ./...
endif
	@rm -rf vendor

