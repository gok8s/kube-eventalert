package utils

import (
	"os"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// High enough QPS to fit all expected use cases. QPS=0 is not set here, because
	// client code is overriding it.
	defaultQPS = 1e6
	// High enough Burst to fit all expected use cases. Burst=0 is not set here, because
	// client code is overriding it.
	defaultBurst = 1e6

	fakeCertificate = "default-fake-certificate"
)

// GetClientOutOfCluster returns a k8s clientset to the request from outside of cluster
func GetClientOutOfCluster(apiServerHost, kubeconfigPath string) kubernetes.Interface {
	config, err := buildOutOfClusterConfig(apiServerHost, kubeconfigPath)
	if err != nil {
		logrus.Fatalf("Can not get kubernetes config: %v", err)
	}

	config.Burst = defaultBurst
	config.QPS = defaultQPS
	config.ContentType = "application/vnd.kubernetes.protobuf"

	clientset, err := kubernetes.NewForConfig(config)

	return clientset
}

func buildOutOfClusterConfig(apiServerHost, kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("KUBECONFIG")
		if kubeconfigPath == "" {
			kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
		}
	}
	if apiServerHost == "" {
		apiServerHost := os.Getenv("APISERVERHOST")
		if apiServerHost == "" {
			apiServerHost = os.Getenv("HOME") + "/.kube/config"
		}
	}
	return clientcmd.BuildConfigFromFlags(apiServerHost, kubeconfigPath)
}

// GetClient returns a k8s clientset to the request from inside of cluster
func GetClient() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.Fatalf("Can not get kubernetes config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("Can not create kubernetes client: %v", err)
	}
	return clientset
}
