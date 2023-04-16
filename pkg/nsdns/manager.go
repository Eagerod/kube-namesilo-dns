package nsdns

import (
	"fmt"
)

import (
	apinetworkingv1 "k8s.io/api/networking/v1"
)

type DnsManager struct {
	BareDomainName     string
	TargetIngressClass string
}

func NewDnsManager(domainName, ingressClass string) (*DnsManager, error) {
	if domainName == "" {
		return nil, fmt.Errorf("must provide a domain name to target DNS record updates")
	}

	if ingressClass == "" {
		return nil, fmt.Errorf("must provide an ingress class to generate DNS records")
	}

	dm := DnsManager{
		domainName,
		ingressClass,
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
