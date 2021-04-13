.PHONY: local
local:
	GOOS=linux go build ../cmd/main.go
	docker build -f Dockerfile.dev -t diago .

remove:
	- kubectl delete sts diago --namespace=diago
	- kubectl delete po -l group=test-worker --namespace=diago

deploy:
	kubectl apply -k ../manifests/

do: remove local deploy

logs:
	kubectl logs diago-0 -f --namespace=diago

watch:
	kubectl get po -n diago -w

create-flow:
	./test/create.sh

flow:
	./test/test.sh