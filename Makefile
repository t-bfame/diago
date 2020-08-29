.PHONY:	build run

build: remove
	GOOS=linux go build cmd/main.go
	docker build -f build/package/Dockerfile -t diago .
	kubectl apply -f deployments/deploy.yaml
	kubectl get po

remove:
	- kubectl delete sts diago
	- kubectl delete svc diago-0
	- kubectl delete po -l group=diago-worker

run:
	kubectl apply -f deployments/deploy.yaml
	kubectl get po

logs:
	kubectl logs diago-0 -f


test:
	./test.sh

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
