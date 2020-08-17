.PHONY:	build run

build:
	GOOS=linux go build cmd/main.go
	docker build -f build/package/Dockerfile -t diago .
	kubectl delete sts diago

remove:
	kubectl delete sts diago
	kubectl delete po diago-worker-6fbbd7

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
		api/proto/worker.proto
