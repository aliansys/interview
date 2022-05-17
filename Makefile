SHELL = /bin/bash
.DEFAULT_GOAL := test

.PHONY: all dep vet test bench docker_prepare docker_start docker_stop docker_cleanup run docker_rmi events_count request run_fake

dep:
	go get -d ./...
	go mod tidy -v

vet:
	go vet ./...

test:
	go test -race -cover ./...

bench:
	go test -run=XXX -benchmem -bench=. ./...

CREATE_DB = $(shell cat './storage/clickhouse/migrations/01_createdatabase.sql')
CREATE_TABLE = $(shell cat './storage/clickhouse/migrations/02_createtable.sql')
docker_prepare:
	@echo 'pulling a ch-server image'
	@docker pull yandex/clickhouse-server:21.3.20.1
	@echo 'creating default container'
	@docker run -d --name interview-clickhouse-server -p 127.0.0.1:8124:8123 -p 127.0.0.1:9001:9000 --ulimit nofile=262144:262144 yandex/clickhouse-server:21.3.20.1
	@echo 'pulling a ch-client image'
	@docker pull yandex/clickhouse-client:21.3.20.1
	sleep 1
	@echo 'running migrations from ./storage/clickhouse/migrations'
	@docker run --rm --link interview-clickhouse-server:clickhouse-server yandex/clickhouse-client:21.3.20.1 --host clickhouse-server --port 9000 --query="$(CREATE_DB)"
	@docker run --rm --link interview-clickhouse-server:clickhouse-server yandex/clickhouse-client:21.3.20.1 --host clickhouse-server --port 9000 --database interview --query='$(CREATE_TABLE)'
	@docker stop interview-clickhouse-server

docker_start:
	@docker start interview-clickhouse-server
	@echo 'interview-clickhouse-server running on port 9001'

docker_stop:
	docker stop interview-clickhouse-server

docker_cleanup:
	docker stop interview-clickhouse-server
	docker rm interview-clickhouse-server

docker_rmi:
	docker rmi yandex/clickhouse-server
	docker rmi yandex/clickhouse-client

events_count:
	@docker run --rm --link interview-clickhouse-server:clickhouse-server yandex/clickhouse-client:21.3.20.1 --host clickhouse-server --port 9000 --query="select count(*) from interview.events"

run: docker_start
	@go run main.go || true

run_fake:
	@go run main.go -no-ch || true

request:
	curl -d 'events={"client_time":"2020-12-01 23:59:00", "device_id":"0287D9AA-4ADF-4B37-A60F-3E9E645C821E", "device_os":"iOS 13.5.1", "session":"ybuRi8mAUypxjbxQ", "sequence":1, "event":"app_start", "param_int":0, "param_str":"some text"}' -X POST localhost:3000/v1/events
