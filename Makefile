.PHONY:	build run

docker:
	docker build -t diago .

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

test:
	go test -v -coverprofile=coverage.out ./...
