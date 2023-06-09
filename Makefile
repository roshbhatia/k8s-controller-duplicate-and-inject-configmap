.PHONY: build-and-load-docker-image deploy clean create-kind-cluster delete-kind-cluster reset-kind-cluster

DOCKER_IMAGE_NAME := env-injector-controller:local

build:
	go build -o ./bin/env-injector-controller ./cmd/main/main.go

build-and-load-docker-image:
	docker build -t $(DOCKER_IMAGE_NAME) .
	kind load docker-image $(DOCKER_IMAGE_NAME) --name kind

deploy:
	kubectl apply -f dev/manifests/controller-deployment.yaml --context kind-kind
	kubectl apply -f dev/manifests/env-configmap-to-inject.yaml --context kind-kind
	kubectl apply -f dev/manifests/pod-with-annotation.yaml --context kind-kind

clean:
	kubectl delete -f dev/manifests/controller-deployment.yaml --context kind-kind
	kubectl delete -f dev/manifests/pod-with-annotation.yaml --context kind-kind
	kubectl delete -f dev/manifests/env-configmap-to-inject.yaml --context kind-kind
	docker rmi $(DOCKER_IMAGE_NAME)

create-kind-cluster:
	kind create cluster --config dev/kind/cluster-config.yaml

delete-kind-cluster:
	kind delete cluster --name kind

reset-kind-cluster: delete-kind-cluster create-kind-cluster
