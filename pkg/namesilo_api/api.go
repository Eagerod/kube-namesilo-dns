package namesilo_api

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const DefaultApiURLPrefix string = "https://www.namesilo.com/api/"

type NamesiloApi struct {
	apiKey    string
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
		Detail          string           `xml:"detail"`
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
		apiKey:    apiKey,
		apiPrefix: DefaultApiURLPrefix,
	}
}

func NewNamesiloApiWithServer(apiKey, apiPrefix string) *NamesiloApi {
	return &NamesiloApi{
		apiKey:    apiKey,
		apiPrefix: apiPrefix,
	}
}

func (ns *NamesiloApi) ListDNSRecords(domain string) ([]ResourceRecord, error) {
	reqUrl, err := url.Parse(fmt.Sprintf("%s/%s", ns.apiPrefix, "dnsListRecords"))
	if err != nil {
		return nil, err
	}

	reqQuery := reqUrl.Query()
	reqQuery.Add("version", "1")
	reqQuery.Add("type", "xml")
	reqQuery.Add("key", ns.apiKey)
	reqQuery.Add("domain", domain)

	reqUrl.RawQuery = reqQuery.Encode()

	response, err := http.Get(reqUrl.String())
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

	if ldrr.Reply.Detail == "success" {
		return ldrr.Reply.ResourceRecords, nil
	}

	return nil, fmt.Errorf("namesilo domain list failed with: %s", ldrr.Reply.Detail)
}

func (ns *NamesiloApi) UpdateDNSRecord(domain, host, id, value string, ttl int) error {
	reqUrl, err := url.Parse(fmt.Sprintf("%s/%s", ns.apiPrefix, "dnsUpdateRecord"))
	if err != nil {
		return err
	}

	reqQuery := reqUrl.Query()
	reqQuery.Add("version", "1")
	reqQuery.Add("type", "xml")
	reqQuery.Add("key", ns.apiKey)
	reqQuery.Add("domain", domain)
	reqQuery.Add("rrid", id)
	reqQuery.Add("rrhost", host)
	reqQuery.Add("rrvalue", value)
	reqQuery.Add("rrttl", strconv.Itoa(ttl))

	reqUrl.RawQuery = reqQuery.Encode()

	response, err := http.Get(reqUrl.String())
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
	reqUrl, err := url.Parse(fmt.Sprintf("%s/%s", ns.apiPrefix, "dnsAddRecord"))
	if err != nil {
		return err
	}

	reqQuery := reqUrl.Query()
	reqQuery.Add("version", "1")
	reqQuery.Add("type", "xml")
	reqQuery.Add("key", ns.apiKey)
	reqQuery.Add("domain", domain)
	reqQuery.Add("rrtype", domainType)
	reqQuery.Add("rrhost", host)
	reqQuery.Add("rrvalue", domain)
	reqQuery.Add("rrttl", strconv.Itoa(ttl))
	reqQuery.Add("rrdistance", "0")

	reqUrl.RawQuery = reqQuery.Encode()

	response, err := http.Get(reqUrl.String())
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
