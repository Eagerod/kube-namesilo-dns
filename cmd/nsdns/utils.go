package cmd

import (
	"context"
	"path/filepath"
)

import (
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/namesilo_api"
	"github.com/Eagerod/kube-namesilo-dns/pkg/nsdns"
)

func GetResourcesFromKubernetesIngresses(domainName string) ([]namesilo_api.ResourceRecord, error) {
	rv := []namesilo_api.ResourceRecord{}

	home := homedir.HomeDir()
	kc := filepath.Join(home, ".kube", "config")
	ctx := context.TODO()
	config, err := clientcmd.BuildConfigFromFlags("", kc)
	if err != nil {
		return rv, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return rv, err
	}

	items, err := clientset.NetworkingV1().Ingresses("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return rv, err
	}

	for _, item := range items.Items {
		if ingressClass, _ := item.Annotations["kubernetes.io/ingress.class"]; ingressClass != "nginx-external" {
			log.Debugf("Skipping ingress %s because it has incorrect ingress class", item.ObjectMeta.Name)
			continue
		}

		nsrr, err := nsdns.NamesiloRecordFromIngress(&item, domainName)
		if err != nil {
			return rv, err
		}

		rv = append(rv, *nsrr)
	}

	return rv, nil
}
