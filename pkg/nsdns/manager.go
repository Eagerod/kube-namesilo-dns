package nsdns

import (
	"fmt"
	"os"
)

import (
	apinetworkingv1 "k8s.io/api/networking/v1"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/namesilo_api"
)

type DnsManager struct {
	BareDomainName     string
	TargetIngressClass string

	Api *namesilo_api.NamesiloApi
}

func NewDnsManager(domainName, ingressClass string) (*DnsManager, error) {
	nsApiKey := os.Getenv("NAMESILO_API_KEY")
	if nsApiKey == "" {
		return nil, fmt.Errorf("failed to find NAMESILO_API_KEY in environment; cannot proceed")
	}

	return NewDnsManagerWithApiKey(domainName, ingressClass, nsApiKey)
}

func NewDnsManagerWithApiKey(domainName, ingressClass, apiKey string) (*DnsManager, error) {
	if domainName == "" {
		return nil, fmt.Errorf("must provide a domain name to target DNS record updates")
	}

	if ingressClass == "" {
		return nil, fmt.Errorf("must provide an ingress class to generate DNS records")
	}

	api := namesilo_api.NewNamesiloApi(domainName, apiKey)

	dm := DnsManager{
		domainName,
		ingressClass,
		api,
	}

	return &dm, nil
}

func (dm *DnsManager) ShouldProcessIngress(ingress *apinetworkingv1.Ingress) bool {
	ic, ok := ingress.Annotations["kubernetes.io/ingress.class"]
	if !ok {
		return false
	}

	return ic == dm.TargetIngressClass
}
