package nsdns

import (
	"fmt"
	"os"
)

import (
	apinetworkingv1 "k8s.io/api/networking/v1"
	log "github.com/sirupsen/logrus"
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

func (dm *DnsManager) HandleIngressExists(ingress *apinetworkingv1.Ingress, cache *DnsManagerCache) error {
	if !dm.ShouldProcessIngress(ingress) {
		return nil
	}

	record, err := NamesiloRecordFromIngress(ingress, dm.BareDomainName, cache.CurrentIpAddress)
	if err != nil {
		return err
	}

	for _, r := range cache.CurrentRecords {
		if record.Type == r.Type && record.Host == r.Host {
			if record.EqualsRecord(r) {
				log.Debugf("Record %s:%s already up to date", record.Type, record.Host)
				return nil
			}

			record.RecordId = r.RecordId
			log.Debugf("Updating record %s:%s with value %s", record.Type, record.Host, record.Value)
			return dm.Api.UpdateDNSRecord(*record)
		}
	}

	log.Debugf("Creating new record %s:%s with value %s", record.Type, record.Host, record.Value)
	return dm.Api.AddDNSRecord(*record)
}

func (dm *DnsManager) HandleIngressDeleted(ingress *apinetworkingv1.Ingress, cache *DnsManagerCache) error {
	if !dm.ShouldProcessIngress(ingress) {
		return nil
	}

	record, err := NamesiloRecordFromIngress(ingress, dm.BareDomainName, cache.CurrentIpAddress)
	if err != nil {
		return err
	}

	for _, r := range cache.CurrentRecords {
		if record.Type == r.Type && record.Host == r.Host {
			log.Infof("Deleting resource record %s", r.RecordId)
			return dm.Api.DeleteDNSRecord(r)
		}
	}

	return fmt.Errorf("failed to find record: %s:%s", record.Type, record.Host)
}
