.PHONY:	build run

BRANCH := $(shell git branch --show-current)

docker:
	docker build -t tbfame/diago:${BRANCH} .

push:
	docker push tbfame/diago:${BRANCH}

PROTOC := protoc

packr:
	GOOS=linux GO111MODULE=on packr2 --ignore-imports

.PHONY: ui
ui:
	cd ui && npm install
	cd ui && npm run build
	mv ui/build dist

package: ui packr

crd-gen:
	controller-gen object paths=./api/v1alpha1/workergroup.go

.PHONY: proto
proto:

	$(PROTOC) \
		--go_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		--go-grpc_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		idl/proto/worker.proto

	$(PROTOC) \
		--go_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		--go-grpc_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		idl/proto/aggregator.proto

test:
	go test -v -coverprofile=coverage.out ./...
