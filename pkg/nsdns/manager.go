package nsdns

import (
	"fmt"
	"os"
	"sync"
)

import (
	log "github.com/sirupsen/logrus"
	apinetworkingv1 "k8s.io/api/networking/v1"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/namesilo_api"
)

type DnsManager struct {
	BareDomainName     string
	TargetIngressClass string

	Api namesilo_api.NamesiloApi

	cacheLock *sync.Mutex
	cache *DnsManagerCache
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
		&sync.Mutex{},
		NewDnsManagerCache(),
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

func (dm *DnsManager) HandleIngressExists(ingress *apinetworkingv1.Ingress) error {
	if !dm.ShouldProcessIngress(ingress) {
		return nil
	}

	record, err := NamesiloRecordFromIngress(ingress, dm.BareDomainName, dm.cache.CurrentIpAddress)
	if err != nil {
		return err
	}

	for _, r := range dm.cache.CurrentRecords {
		if record.Type == r.Type && record.Host == r.Host {
			if record.EqualsRecord(r) {
				log.Debugf("Record %s:%s already up to date", record.Type, record.Host)
				return nil
			}

			record.RecordId = r.RecordId
			log.Debugf("Updating record %s:%s with value %s", record.Type, record.Host, record.Value)
			if err := dm.Api.UpdateDNSRecord(*record); err != nil {
				return err
			}

			return dm.UpdateCache()
		}
	}

	log.Debugf("Creating new record %s:%s with value %s", record.Type, record.Host, record.Value)
	if err := dm.Api.AddDNSRecord(*record); err != nil {
		return err
	}
	return dm.UpdateCache()
}

func (dm *DnsManager) HandleIngressDeleted(ingress *apinetworkingv1.Ingress) error {
	if !dm.ShouldProcessIngress(ingress) {
		return nil
	}

	record, err := NamesiloRecordFromIngress(ingress, dm.BareDomainName, dm.cache.CurrentIpAddress)
	if err != nil {
		return err
	}

	for _, r := range dm.cache.CurrentRecords {
		if record.Type == r.Type && record.Host == r.Host {
			log.Infof("Deleting resource record %s", r.RecordId)
			if err := dm.Api.DeleteDNSRecord(r); err != nil {
				return err
			}

			return dm.UpdateCache()
		}
	}

	return fmt.Errorf("failed to find record: %s:%s", record.Type, record.Host)
}

func (dm *DnsManager) UpdateCache() error {
	dm.cacheLock.Lock()
	defer dm.cacheLock.Unlock()

	if err := UpdateCachedRecords(dm.cache, dm.Api); err != nil {
		return err
	}

	log.Debugf("Received %d records from Namesilo", len(dm.cache.CurrentRecords))

	if err := UpdateIpAddress(dm.cache); err != nil {
		return err
	}

	return nil
}
