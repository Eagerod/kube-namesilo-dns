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
	reqValues := url.Values{}

	reqValues.Add("domain", domain)

	reqUrl, err := ns.apiActionWithValues("dnsListRecords", &reqValues)
	if err != nil {
		return nil, err
	}

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
	reqValues := url.Values{}

	reqValues.Add("domain", domain)
	reqValues.Add("rrid", id)
	reqValues.Add("rrhost", host)
	reqValues.Add("rrvalue", value)
	reqValues.Add("rrttl", strconv.Itoa(ttl))

	reqUrl, err := ns.apiActionWithValues("dnsUpdateRecord", &reqValues)
	if err != nil {
		return err
	}

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
	reqValues := url.Values{}

	reqValues.Add("domain", domain)
	reqValues.Add("rrtype", domainType)
	reqValues.Add("rrhost", host)
	reqValues.Add("rrvalue", domain)
	reqValues.Add("rrttl", strconv.Itoa(ttl))
	reqValues.Add("rrdistance", "0")

	reqUrl, err := ns.apiActionWithValues("dnsAddRecord", &reqValues)
	if err != nil {
		return err
	}

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

func (ns *NamesiloApi) apiActionWithValues(action string, values *url.Values) (*url.URL, error) {
	reqUrl, err := url.Parse(fmt.Sprintf("%s/%s", ns.apiPrefix, action))
	if err != nil {
		return nil, err
	}

	newValues := *values

	newValues.Add("version", "1")
	newValues.Add("type", "xml")
	newValues.Add("key", ns.apiKey)

	reqUrl.RawQuery = newValues.Encode()

	return reqUrl, nil
}
