package main

import (
	"context"
	"flag"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
			if configMapName, ok := newPod.Annotations[annotationKey]; ok {
				klog.Info("Found new pod with annotation")
				klog.Info("Injecting environment variables from ConfigMap")
				injectConfigMapIntoEnv(newPod, configMapName, clientset)
			}
		},
	})

	shouldExit := make(chan struct{})
	defer close(shouldExit)
	factory.Start(shouldExit)
	<-shouldExit
}

func injectConfigMapIntoEnv(pod *corev1.Pod, configMapName string, clientset *kubernetes.Clientset) {
	ctx := context.Background()

	configMap, err := clientset.CoreV1().ConfigMaps(pod.Namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Error fetching ConfigMap %s: %v\n", configMapName, err)
		return
	}
	klog.Infof("Found ConfigMap %s: %v", configMapName, configMap.Data)

	klog.Info("Injecting environment variables into pod's container")

	klog.Infof("Pod before update: %v", pod)

		// Map ConfigMap key value pairs into something we can inject.
	envVars := []corev1.EnvVar{}
	for k, v := range configMap.Data {
		envVar := corev1.EnvVar{
			Name:  k,
			Value: v,
		}
		envVars = append(envVars, envVar)
	}

	// Assume pod is of size one, as that's what we're testing with.
	pod.Spec.Containers[0].Env = envVars

	klog.Infof("Updated pod: %v", pod)
}
