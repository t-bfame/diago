.PHONY:	build run

build:
	GOOS=linux go build cmd/main.go
	docker build -f build/package/Dockerfile -t diago .
	kubectl delete sts diago

run:
	kubectl apply -f deployments/deploy.yaml
	kubectl get po -w
	kubectl logs diago-0 -f

.PHONY: local
local:
	go build -o ser cmd/server/lol.go
	go build -o cle cmd/server/cle.go

PROTOC := protoc

.PHONY: proto
proto:

	$(PROTOC) \
		--go_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		--go-grpc_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		api/proto/lol.proto
