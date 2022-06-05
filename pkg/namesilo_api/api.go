package namesilo_api

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

const NamesiloApiURLPrefix string = "https://www.namesilo.com/api/"

type NamesiloApi struct {
	apiKey string
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

type DNSUpdateRecordsResponse struct {
	XMLName xml.Name `xml:"namesilo"`
	Reply   struct {
		XMLName xml.Name `xml:"reply"`
		Detail  string   `xml:"detail"`
	}
}

func NewNamesiloApi(apiKey string) *NamesiloApi {
	return &NamesiloApi{
		apiKey: apiKey,
	}
}

func (ns *NamesiloApi) ListDNSRecords(domain string) ([]ResourceRecord, error) {
	url := fmt.Sprintf("%s/%s?version=1&type=xml&key=%s&domain=%s", NamesiloApiURLPrefix, "dnsListRecords", ns.apiKey, domain)
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
	url := fmt.Sprintf("%s/%s?version=1&type=xml&key=%s&domain=%s&rrid=%s&rrhost=%s&rrvalue=%s&rrttl=%d", NamesiloApiURLPrefix, "dnsUpdateRecord", ns.apiKey, domain, id, host, value, ttl)
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

	return fmt.Errorf("Namesilo domain update failed with: %s", durr.Reply.Detail)
}
