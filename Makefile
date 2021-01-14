.PHONY:	build run

.PHONY: local
local:
	GOOS=linux go build cmd/main.go
	eval $(minikube docker-env)
	docker build -f Dockerfile.dev -t diago .

docker:
	docker build -t diago .

remove:
	- kubectl delete sts diago --namespace=diago
	- kubectl delete po -l group=test-worker --namespace=diago

deploy:
	kubectl apply -k manifests/

do: local remove deploy

logs:
	kubectl logs diago-0 -f --namespace=diago

PROTOC := protoc

.PHONY: proto
proto:

	$(PROTOC) \
		--go_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		--go-grpc_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		idl/proto/worker.proto

crd-gen:
	controller-gen object paths=./api/v1alpha1/workergroup.go

# test:
# 	./test.sh

test:
	go test -v -coverprofile=coverage.out ./...

flow:
	./test/test.sh