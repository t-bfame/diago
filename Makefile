proto:
	@ if ! which protoc > /dev/null; then \
		echo "error: protoc not installed" >&2; \
		exit 1; \
	fi
	@ if ! which protoc-gen-go > /dev/null; then \
		echo "error: protoc-gen-go not installed" >&2; \
		exit 1; \
	fi

	@ echo Compiling Protobufs
	@ for file in $$(git ls-files '*.proto'); do \
		protoc \
		--go_out=Mgrpc/service_config/service_config.proto=/internal/proto/grpc_service_config:. \
		--go-grpc_out=Mgrpc/service_config/service_config.proto=/internal/proto/grpc_service_config:. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		$$file; \
	done

build:
	@ go install .

run:
	@ go run main
