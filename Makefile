ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HAS_MIGRATE:=$(shell command -v $(GOPATH)/bin/migrate;)
HAS_POSTGRES_RUNNING:=$(shell docker ps -aqf "name=vinylretailers-postgres*";)

.PHONY: docker
docker:
	# removed all images ad containers, and creates docker network bridge vinylretailers
	@docker ps -aqf "name=vinylretailers-*" | xargs docker rm -f
	@docker images -q | xargs docker image rm
	@docker network rm vinylretailers
	@docker network create --driver bridge -o "com.docker.network.bridge.host_binding_ipv4"="0.0.0.0" -o "com.docker.network.bridge.enable_ip_masquerade"="true" -o "com.docker.network.bridge.enable_icc"="true"  vinylretailers
	make postgres
	make pgadmin
	make redis
	make scanner

.PHONY: redis
redis:
	# grabbing the redis image for docker and starting a container instance
	@docker ps -aqf "name=vinylretailers-redis*" | xargs docker rm -f
	@docker images "redis" -q | xargs docker image rm
	@docker pull redis:latest
	@docker run --name vinylretailers-redis --network=vinylretailers -p 127.0.0.1:6379:6379 --expose=6379 -d redis

.PHONY: postgres
postgres:
	# grabbing the postgres image for docker and starting a container instance
	# need to make this stop conditional on whether postgres container is running
ifdef HAS_POSTGRES_RUNNING
	@echo CLEANING UP OLD POSTGRES INSTANCE
	@docker stop vinylretailers-postgres
	@docker ps -aqf "name=vinylretailers-postgres*" | xargs docker rm -f
	@docker images "postgres" -q | xargs docker image rm
	@docker volume rm pgdata
	rm -rf /Users/gavin/vinylretailers_data
endif
	@docker ps -aqf "name=vinylretailers-postgres*" | xargs docker rm -f
	@docker images "postgres" -q | xargs docker image rm
	@docker pull postgres
	mkdir ~/vinylretailers_data
	@docker volume create --name pgdata --opt type=none --opt device=/Users/gavin/vinylretailers_data --opt o=bind
	@docker run --name vinylretailers-postgres --network=vinylretailers -e POSTGRES_USER="vinylretailers" -e POSTGRES_PASSWORD="vinylretailers" -p 5400:5432 -v pgdata:/var/lib/postgresql/data  -d postgres
	@docker start vinylretailers-postgres
	# we need to wait for a few seconds for the db to be up before runnign the migration
	make migrate

.PHONY: pgadmin
pgadmin:
	# installs and runs the pgAdmin web interface
	@docker ps -aqf "name=vinylretailers-pgadmin*" | xargs docker rm -f
	@docker images "dpage/pgadmin4" -q | xargs docker image rm
	@docker pull dpage/pgadmin4:latest
	@docker run --name vinylretailers-pgadmin --network=vinylretailers -p 80:80 -e 'PGADMIN_DEFAULT_EMAIL=gturner.au@gmail.com' -e 'PGADMIN_DEFAULT_PASSWORD=vinylretailers' -e PGADMIN_SERVER_JSON_FILE='./util/postgres/pgadmin_servers.json' -d dpage/pgadmin4

.PHONY: scanner
scanner:
	# building the retailer scanner image for docker and starting a container instance
	@docker ps -aqf "name=vinylretailers-scanner*" | xargs docker rm -f
	@docker images "vinylretailers-scanner" -q | xargs docker image rm
	@docker build -t vinylretailers-scanner -f cmd/scanner/Dockerfile.scanner .
	@docker run --name vinylretailers-scanner --network=vinylretailers -d vinylretailers-scanner

.PHONY: scheduler
scheduler:
	# building the retailer scheduler image for docker and starting a container instance
	@docker ps -aqf "name=vinylretailers-scheduler*" | xargs docker rm -f
	@docker images "vinylretailers-scheduler" -q | xargs docker image rm
	@docker build -t vinylretailers-scheduler -f cmd/scanner/Dockerfile.scheduler .
	@docker run --name vinylretailers-scheduler --network=vinylretailers -d vinylretailers-scheduler

.PHONY: reporter
scheduler:
	# building the retailer reporter image for docker and starting a container instance
	@docker ps -aqf "name=vinylretailers-reporter*" | xargs docker rm -f
	@docker images "vinylretailers-reporter" -q | xargs docker image rm
	@docker build -t vinylretailers-reporter -f cmd/scanner/Dockerfile.reporter .
	@docker run --name vinylretailers-reporter --network=vinylretailers -d vinylretailers-reporter

.PHONY: migrate
migrate:
	# run this the first time after bringing up the postgres database to initialise the tables and lookup data
ifndef HAS_MIGRATE
	go get -u -d github.com/mattes/migrate/cli github.com/lib/pq
	go build -tags 'postgres' -v -o $(GOPATH)/bin/migrate github.com/mattes/migrate/cli
endif
	@docker ps -aqf "name=vinylretailers-scanner" | xargs docker rm -f
	$(GOPATH)/bin/migrate -path ./db/migrations -database postgres://vinylretailers:vinylretailers@localhost:5400/vinylretailers?sslmode=disable up

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

