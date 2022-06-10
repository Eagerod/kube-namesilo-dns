package nsdns

import (
	networkingv1 "k8s.io/api/networking/v1"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/namesilo_api"
)

func NamesiloRecordFromIngress(ingress *networkingv1.Ingress, domainName, ip string) (*namesilo_api.ResourceRecord, error) {
	rr := namesilo_api.ResourceRecord{}
	rr.Host = ingress.Spec.Rules[0].Host
	rr.TTL = 7207

	if rr.Host == domainName {
		rr.Type = "A"
		rr.Value = ip
	} else {
		rr.Type = "CNAME"
		rr.Value = domainName
	}

	return &rr, nil
}
