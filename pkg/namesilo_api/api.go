package namesilo_api

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

const DefaultApiURLPrefix string = "https://www.namesilo.com/api/"

type NamesiloApi struct {
	apiKey string
	apiPrefix string
}

type ResourceRecord struct {
	XMLName  xml.Name `xml:"resource_record"`
	RecordId string   `xml:"record_id"`
	Type     string   `xml:"type"`
	Host     string   `xml:"host"`
	Value    string   `xml:"value"`
	TTL      int      `xml:"ttl"`
	Distance int      `xml:"distance"`
}

type ListDNSRecordsResponse struct {
	XMLName xml.Name `xml:"namesilo"`
	Reply   struct {
		XMLName         xml.Name         `xml:"reply"`
		ResourceRecords []ResourceRecord `xml:"resource_record"`
	}
}

type DNSAddRecordsResponse struct {
	XMLName xml.Name `xml:"namesilo"`
	Reply   struct {
		XMLName xml.Name `xml:"reply"`
		Detail  string   `xml:"detail"`
	}
}

type DNSUpdateRecordsResponse DNSAddRecordsResponse

func NewNamesiloApi(apiKey string) *NamesiloApi {
	return &NamesiloApi{
		apiKey: apiKey,
		apiPrefix: DefaultApiURLPrefix,
	}
}

func NewNamesiloApiWithServer(apiKey, apiPrefix string) *NamesiloApi {
	return &NamesiloApi{
		apiKey: apiKey,
		apiPrefix: apiPrefix,
	}
}

func (ns *NamesiloApi) ListDNSRecords(domain string) ([]ResourceRecord, error) {
	url := fmt.Sprintf("%s/%s?version=1&type=xml&key=%s&domain=%s", ns.apiPrefix, "dnsListRecords", ns.apiKey, domain)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var ldrr ListDNSRecordsResponse
	if err := xml.Unmarshal(body, &ldrr); err != nil {
		return nil, err
	}

	return ldrr.Reply.ResourceRecords, nil
}

func (ns *NamesiloApi) UpdateDNSRecord(domain, host, id, value string, ttl int) error {
	url := fmt.Sprintf("%s/%s?version=1&type=xml&key=%s&domain=%s&rrid=%s&rrhost=%s&rrvalue=%s&rrttl=%d", ns.apiPrefix, "dnsUpdateRecord", ns.apiKey, domain, id, host, value, ttl)
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var durr DNSUpdateRecordsResponse
	if err := xml.Unmarshal(body, &durr); err != nil {
		return err
	}

	if durr.Reply.Detail == "success" {
		return nil
	}

	return fmt.Errorf("namesilo domain update failed with: %s", durr.Reply.Detail)
}

func (ns *NamesiloApi) AddDNSRecord(domain, domainType, host, value string, ttl int) error {
	url := fmt.Sprintf("%s/%s?version=1&type=xml&key=%s&domain=%s&rrtype=%s&rrhost=%s&rrvalue=%s&rrttl=%d&rrdistance=0", ns.apiPrefix, "dnsAddRecord", ns.apiKey, domain, domainType, host, domain, ttl)
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var durr DNSAddRecordsResponse
	if err := xml.Unmarshal(body, &durr); err != nil {
		return err
	}

	if durr.Reply.Detail == "success" {
		return nil
	}

	return fmt.Errorf("namesilo domain add failed with: %s", durr.Reply.Detail)
}
