ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HAS_MIGRATE := $(shell command -v migrate;)

.PHONY: docker
docker:
	# removed all images ad containres, and creates docker network bridge vinylretailers
	@docker ps -aqf "name=vinylretailers-*" | xargs docker rm -f
	@docker images -q | xargs docker image rm
	@docker network rm vinylretailers
	@docker network create --driver bridge -o "com.docker.network.bridge.host_binding_ipv4"="0.0.0.0" -o "com.docker.network.bridge.enable_ip_masquerade"="true" -o "com.docker.network.bridge.enable_icc"="true"  vinylretailers
	make redis
	make scanner

.PHONY: redis
redis:
	# grabbing the redis image for docker and starting a container instance
	@docker ps -aqf "name=vinylretailers-redis*" | xargs docker rm -f
	@docker images "redis" -q | xargs docker image rm
	@docker pull redis:latest
	@docker run --name vinylretailers-redis --network=vinylretailers -p 6379:6379 --expose=6379 -d redis

.PHONY: scanner
scanner:
	# building the retailer scanner image for docker and starting a container instance
	@docker ps -aqf "name=vinylretailers-scanner*" | xargs docker rm -f
	@docker images "vinylretailers-scanner" -q | xargs docker image rm
	@docker build -t vinylretailers-scanner -f cmd/scanner/Dockerfile.scanner .
	@docker run --name vinylretailers-scanner --network=vinylretailers -d vinylretailers-scanner



.PHONY: generate
generate:
	# TOP TIP: "make generate pkg=[package name]" runs go generate for a specific package
	@go mod vendor
	@docker build -t vinylretailers-makegenerate - < Dockerfile.generate
ifdef pkg
	@docker run --rm -it -e GO111MODULE=off -e GOPATH=$(HOME)/go -v $(HOME)/go:$(HOME)/go:delegated -v $(ROOT_DIR):$(HOME)/go/src/github.com/gavinturner/vinylretailers:delegated -w $(HOME)/go/src/github.com/gavinturner/vinylretailers vinylretailers-makegenerate go generate -mod vendor -x ./$(pkg)/...
else
	@docker run --rm -it -e GO111MODULE=off -e GOPATH=$(HOME)/go -v $(HOME)/go:$(HOME)/go:delegated -v $(ROOT_DIR):$(HOME)/go/src/github.com/gavinturner/vinylretailers:delegated -w $(HOME)/go/src/github.com/gavinturner/vinylretailers vinylretailers-makegenerate go generate -mod vendor -x ./...
endif
	@rm -rf vendor

.PHONY: deps
deps:
ifndef HAS_MIGRATE
	go get -u -d github.com/mattes/migrate/cli github.com/lib/pq
	go build -tags 'postgres' -v -o $(GOPATH)/bin/migrate github.com/mattes/migrate/cli
endif
	@docker ps -aqf "name=vendorrisk_*" | xargs docker rm -f
	@docker ps -aqf "name=vendor_risk_*" | xargs docker rm -f
	@docker-compose -f docker-compose-deps.yaml up -d
	@until migrate -path ./survey/surveydb/migrations -database postgres://cyberrisk:@127.0.0.1:6000/cyberrisk?sslmode=disable up; do echo "database not running yet or error occuring"; sleep 2; done;

