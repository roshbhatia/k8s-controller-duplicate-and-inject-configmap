# k8s-controller-duplicate-and-inject-configmap

A sample k8s controller that injects environment variables from a ConfigMap into a new Pod after admission based on an existing pod with the annotation(`duplicate-and-duplicate-and-inject-configmap: $CONFIG_MAP_NAME`).

## Running the Project

This repo includes a Makefile; a few of the relevant targets are:

- `create-kind-cluster`: Creates a local Kind k8s cluster for local development.
- `build-and-load-docker-image`: Builds the controller's docker image and loads it into Kind.
- `deploy`: Applies a deployment for the controller, associated ServiceAccount and ClusterRole resources, and a sample ConfigMap to inject and a Pod with the annotation.
- `clean`: Deletes resources and local docker image.

## Limitations

I built this as an initial foray into building with k8s instead of just against it. As a result, it really doesn't do anything complex.

The controller only does the following:

- Watches for new pods with the annotation
- Duplicates said pod, but with the contents of the specified ConfigMap.

It doesn't do any sort of validation in the ConfigMap, or checks for updated pods.
