package main

import (
	"flag"
	"time"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	annotationKey = "inject-configmap"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	klog.Info("Initializing controller")

	klog.InitFlags(nil)
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Error(err, "Error building kubeconfig")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	klog.Infof("Parsed config: %v", cfg)

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}

	factory := informers.NewSharedInformerFactory(clientset, 30*time.Second)
	podInformer := factory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			newPod := obj.(*corev1.Pod)
			if _, ok := newPod.Annotations[annotationKey]; ok {
				klog.Info("Found new pod with annotation")
			}
		},
		UpdateFunc: func(_, newObj interface{}) {
			newPod := newObj.(*corev1.Pod)

			if _, ok := newPod.Annotations[annotationKey]; ok {
				klog.Info("Found updated pod with annotation")
			}
		},
	})

	shouldExit := make(chan struct{})
	defer close(shouldExit)
	factory.Start(shouldExit)
	<-shouldExit
}
