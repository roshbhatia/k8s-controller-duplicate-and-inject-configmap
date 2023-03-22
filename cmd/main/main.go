package main

import (
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	fmt.Println("Loading k8s config")
	kubeconfig := os.Getenv("KUBECONFIG")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(config)
}
