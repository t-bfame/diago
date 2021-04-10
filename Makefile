.PHONY:	build run

.PHONY: local
local:
	GOOS=linux go build cmd/main.go
	docker build -f Dockerfile.dev -t diago .

build-local-ui:
	cd ui && npm install
	cd ui && npm run build
	mv ui/build dist

local-ui: build-local-ui ui local

docker:
	docker build -t diago .

remove:
	- kubectl delete sts diago --namespace=diago
	- kubectl delete po -l group=test-worker --namespace=diago

deploy:
	kubectl apply -k manifests/

do: remove local deploy

logs:
	kubectl logs diago-0 -f --namespace=diago

watch:
	kubectl get po -n diago -w

PROTOC := protoc

.PHONY: proto
proto:

	$(PROTOC) \
		--go_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		--go-grpc_out=Mgrpc/service_config/service_config.proto=/proto-gen/api:. \
		idl/proto/worker.proto

.PHONY: ui
ui:
	GOOS=linux GO111MODULE=on packr2 --ignore-imports

crd-gen:
	controller-gen object paths=./api/v1alpha1/workergroup.go

# test:
# 	./test.sh

test:
	go test -v -coverprofile=coverage.out ./...

create-flow:
	./test/create.sh

flow:
	./test/test.sh