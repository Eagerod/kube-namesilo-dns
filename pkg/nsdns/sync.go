package nsdns

import (
	"fmt"
	"strings"
)

import (
	networkingv1 "k8s.io/api/networking/v1"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/namesilo_api"
)

func NamesiloRecordFromIngress(ingress *networkingv1.Ingress, domainName string) (*namesilo_api.ResourceRecord, error) {
	domainReplacer := fmt.Sprintf(".%s", domainName)
	rr := namesilo_api.ResourceRecord{}
	rr.Host = strings.TrimSuffix(ingress.Spec.Rules[0].Host, domainReplacer)
	rr.TTL = 7207
	rr.Type = "CNAME"
	rr.Value = domainName
	return &rr, nil
}
