package nsdns

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/icanhazip"
	"github.com/Eagerod/kube-namesilo-dns/pkg/namesilo_api"
)

type DnsManagerCache struct {
	CurrentRecords   []namesilo_api.ResourceRecord
	CurrentIpAddress string
}

func NewDnsManagerCache() *DnsManagerCache {
	return &DnsManagerCache{
		[]namesilo_api.ResourceRecord{},
		"",
	}
}

// These aren't implemented as methods on the DnsManagerCache struct, because
// they seem like the kinds of things it shouldn't know how to do on its own.
func UpdateCachedRecords(cache *DnsManagerCache, api *namesilo_api.NamesiloApi) error {
	records, err := api.ListDNSRecords()
	if err != nil {
		return err
	}

	cache.CurrentRecords = records
	return nil
}

func UpdateIpAddress(cache *DnsManagerCache) error {
	ip, err := icanhazip.GetPublicIP()
	if err != nil {
		return err
	}

	cache.CurrentIpAddress = ip
	return nil
}
