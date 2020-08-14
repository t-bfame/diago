# proto:
# 	@ if ! which protoc > /dev/null; then \
# 		echo "error: protoc not installed" >&2; \
# 		exit 1; \
# 	fi
# 	@ if ! which protoc-gen-go > /dev/null; then \
# 		echo "error: protoc-gen-go not installed" >&2; \
# 		exit 1; \
# 	fi

# 	@ echo Compiling Protobufs
# 	@ for file in $$(git ls-files '*.proto'); do \
# 		protoc \
# 		--go_out=Mgrpc/service_config/service_config.proto=/internal/proto/grpc_service_config:. \
# 		--go-grpc_out=Mgrpc/service_config/service_config.proto=/internal/proto/grpc_service_config:. \
# 		--go_opt=paths=source_relative \
# 		--go-grpc_opt=paths=source_relative \
# 		$$file; \
# 	done

.PHONY:	build run

build:
	GOOS=linux go build cmd/main.go
	docker build -f build/package/Dockerfile -t diago .
	kubectl delete sts diago

run:
	kubectl apply -f deployments/deploy.yaml
	kubectl get po -w
	kubectl logs diago-0 -f
