.PHONY:	build run

build:
	GOOS=linux go build cmd/main.go
	# kubectl apply -f deployments/deploy.yaml
	# kubectl get po

docker:
	docker build -f build/package/Dockerfile -t diago .

remove:
	- kubectl delete sts diago
	- kubectl delete svc diago-0
	- kubectl delete po -l group=diago-worker

run:
	kubectl apply -f deployments/deploy.yaml
	kubectl get po

logs:
	kubectl logs diago-0 -f

.PHONY: local
local:
	go build cmd/main.go

PROTOC := protoc

.PHONY: proto
proto:

	$(PROTOC) \
		--go_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		--go-grpc_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		idl/proto/worker.proto

# test:
# 	./test.sh

test:
	go test -v -coverprofile=coverage.out ./...
