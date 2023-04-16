package cmd

import (
	"context"
	"errors"
	"os"
	"path"
)

import (
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/namesilo_api"
	"github.com/Eagerod/kube-namesilo-dns/pkg/nsdns"
)

type RecordReconciliation struct {
	Add    []namesilo_api.ResourceRecord
	Update []namesilo_api.ResourceRecord
	NoOp   []namesilo_api.ResourceRecord
}

func GetResourcesFromKubernetesIngresses(domainName, ip, ingressClass string) ([]namesilo_api.ResourceRecord, error) {
	rv := []namesilo_api.ResourceRecord{}

	clientset, err := GetKubernetesClientSet()
	if err != nil {
		return rv, err
	}

	ctx := context.TODO()
	items, err := clientset.NetworkingV1().Ingresses("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return rv, err
	}

	for _, item := range items.Items {
		if !ShouldProcessIngress(ingressClass, &item) {
			log.Debugf("Skipping ingress %s because it has incorrect ingress class", item.ObjectMeta.Name)
			continue
		}

		nsrr, err := nsdns.NamesiloRecordFromIngress(&item, domainName, ip)
		if err != nil {
			return rv, err
		}

		rv = append(rv, *nsrr)
	}

	return rv, nil
}

func ShouldProcessIngress(desiredIngressClass string, ingress *networkingv1.Ingress) bool {
	ic, ok := ingress.Annotations["kubernetes.io/ingress.class"]
	if !ok {
		return false
	}
	return ic == desiredIngressClass
}

func ReconcileRecords(existing, new []namesilo_api.ResourceRecord) RecordReconciliation {
	rr := RecordReconciliation{}

	existingByHost := map[string]namesilo_api.ResourceRecord{}
	for _, res := range existing {
		existingByHost[res.Host] = res
	}

	for _, res := range new {
		if r, ok := existingByHost[res.Host]; ok {
			if r.EqualsRecord(res) {
				rr.NoOp = append(rr.NoOp, res)
			} else {
				if r.RecordId != "" {
					res.RecordId = r.RecordId
				}
				rr.Update = append(rr.Update, res)
			}
		} else {
			rr.Add = append(rr.Add, res)
		}
	}

	return rr
}

func RecordMatching(records []namesilo_api.ResourceRecord, record namesilo_api.ResourceRecord) (*namesilo_api.ResourceRecord, error) {
	for _, r := range records {
		if record.EqualsRecord(r) {
			if r.RecordId != "" {
				record.RecordId = r.RecordId
			}
			return &record, nil
		}
	}

	return nil, nil
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
