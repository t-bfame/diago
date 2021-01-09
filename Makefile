.PHONY:	build run

.PHONY: local
local:
	GOOS=linux go build cmd/main.go
	docker build -f Dockerfile.dev -t diago .

docker:
	docker build -t diago .

remove:
	- kubectl delete sts diago
	- kubectl delete svc diago-0
	- kubectl delete po -l group=test-worker

run:
	kubectl apply -f deployments/deploy.yaml
	kubectl get po

do: local remove run

logs:
	kubectl logs diago-0 -f

PROTOC := protoc

.PHONY: proto
proto:

	$(PROTOC) \
		--go_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		--go-grpc_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		idl/proto/worker.proto

crd-gen:
	controller-gen object paths=./api/v1alpha1/worker.go

# test:
# 	./test.sh

test:
	go test -v -coverprofile=coverage.out ./...
