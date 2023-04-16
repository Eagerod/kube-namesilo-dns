package cmd

import (
	"context"
	"errors"
	"os"
	"path"
)

import (
	log "github.com/sirupsen/logrus"
	apinetworkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/namesilo_api"
)

type RecordReconciliation struct {
	Add    []namesilo_api.ResourceRecord
	Update []namesilo_api.ResourceRecord
	NoOp   []namesilo_api.ResourceRecord
}

func GetIngresses(namespace string) ([]apinetworkingv1.Ingress, error) {
	rv := []apinetworkingv1.Ingress{}

	clientset, err := GetKubernetesClientSet()
	if err != nil {
		return rv, err
	}

	ctx := context.TODO()
	items, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return rv, err
	}

	rv = append(rv, items.Items...)

	return rv, nil
}

func GetKubernetesClientSet() (*kubernetes.Clientset, error) {
	if config, err := rest.InClusterConfig(); err == nil {
		return kubernetes.NewForConfig(config)
	} else {
		log.Info(err.Error())
	}

	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = path.Join(homedir.HomeDir(), ".kube", "config")
	}
	if config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath); err == nil {
		return kubernetes.NewForConfig(config)
	} else {
		log.Info(err.Error())
	}

	return nil, errors.New("failed to configure Kubernetes client")
}
