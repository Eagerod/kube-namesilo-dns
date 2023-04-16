package nsdns

import (
	"fmt"
)

type DnsManager struct {
	BareDomainName string
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

