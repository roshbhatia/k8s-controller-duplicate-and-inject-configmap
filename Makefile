.PHONY: build-docker-image load-docker-image deploy clean

DOCKER_IMAGE_NAME := env-injector-controller:local

build:
	go build -o ./bin/env-injector-controller ./cmd/main/main.go

build-docker-image:
	docker build -t $(DOCKER_IMAGE_NAME) .

load-docker-image: build-docker-image
	kind load docker-image $(DOCKER_IMAGE_NAME) --name kind

deploy: load-docker-image
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
