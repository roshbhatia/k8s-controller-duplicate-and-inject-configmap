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
	annotationKey   = "duplicate-and-inject-configmap"
	clonedPodSuffix = "-with-env-injected"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
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

	ctx := context.Background()

	klog.Info("Initialized controller")

	factory := informers.NewSharedInformerFactory(clientset, 30*time.Second)
	podInformer := factory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			if configMapName, ok := pod.Annotations[annotationKey]; ok {
				klog.Info("Found new pod with annotation")

				configMap, err := clientset.CoreV1().ConfigMaps(pod.Namespace).Get(ctx, configMapName, metav1.GetOptions{})
				if err != nil {
					klog.Errorf("Error fetching ConfigMap %s: %v\n", configMapName, err)
					return
				}
				klog.Info("Fetched referenced ConfigMap: %v", configMap)

				newPod, err := clonePodWithModifications(pod, configMap, clientset)
				if err != nil {
					klog.Errorf("Error attempting to clone and modify pod: %v\n", err)
					return
				}
				klog.Infof("Cloned pod definition and injected ConfigMap into env: %v", newPod)

				_, err = clientset.CoreV1().Pods(newPod.Namespace).Create(ctx, newPod, metav1.CreateOptions{
					FieldValidation: "Ignore",
					// This field will be present in the final manifest.
					FieldManager: "env-injector-controller",
				})

				if err != nil {
					klog.Errorf("Error creating new pod: %v\n", err)
					return
				}
				klog.Infof("Created new pod: %v", newPod)
			}
		},
	})

	shouldExit := make(chan struct{})
	defer close(shouldExit)
	factory.Start(shouldExit)
	<-shouldExit
}

func clonePodWithModifications(pod *corev1.Pod, configMap *corev1.ConfigMap, clientset *kubernetes.Clientset) (newPod *corev1.Pod, err error) {
	// Map ConfigMap key value pairs into something we can inject.
	envVars := []corev1.EnvVar{}
	for k, v := range configMap.Data {
		envVar := corev1.EnvVar{
			Name:  k,
			Value: v,
		}
		envVars = append(envVars, envVar)
	}

	newPod = pod.DeepCopy()

	newPod.Name = pod.Name + clonedPodSuffix

	// We're making the assumption that the pod is of size one, as that's what we're testing with.
	newPod.Spec.Containers[0].Env = envVars

	// Reset the resource version as that's set by the API server.
	newPod.SetResourceVersion("")

	// Remove the annotation so the controller doesn't pick up the new pod.
	delete(newPod.Annotations, annotationKey)

	if err != nil {
		klog.Errorf("Error duplicating Pod %s: %v\n", pod.Name, err)
		return nil, err
	}

	return newPod, nil
}
