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
	domain    string
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

func NewNamesiloApi(domain, apiKey string) *NamesiloApi {
	return &NamesiloApi{
		apiKey:    apiKey,
		apiPrefix: DefaultApiURLPrefix,
		domain:    domain,
	}
}

func NewNamesiloApiWithServer(domain, apiKey, apiPrefix string) *NamesiloApi {
	return &NamesiloApi{
		apiKey:    apiKey,
		apiPrefix: apiPrefix,
		domain:    domain,
	}
}

func (ns *NamesiloApi) ListDNSRecords() ([]ResourceRecord, error) {
	reqValues := url.Values{}
	reqUrl, err := ns.apiActionWithValues("dnsListRecords", &reqValues)
	if err != nil {
		return nil, err
	}

	var ldrr ListDNSRecordsResponse
	if err := request(reqUrl, &ldrr); err != nil {
		return nil, err
	} else if ldrr.Reply.Detail != "success" {
		return nil, fmt.Errorf("namesilo domain list failed with: %s", ldrr.Reply.Detail)
	}

	return ldrr.Reply.ResourceRecords, nil
}

func (ns *NamesiloApi) UpdateDNSRecord(domain, host, id, value string, ttl int) error {
	reqValues := url.Values{}

	reqValues.Add("rrid", id)
	reqValues.Add("rrhost", host)
	reqValues.Add("rrvalue", value)
	reqValues.Add("rrttl", strconv.Itoa(ttl))

	reqUrl, err := ns.apiActionWithValues("dnsUpdateRecord", &reqValues)
	if err != nil {
		return err
	}

	var durr DNSUpdateRecordsResponse
	if err := request(reqUrl, &durr); err != nil {
		return err
	} else if durr.Reply.Detail != "success" {
		return fmt.Errorf("namesilo domain update failed with: %s", durr.Reply.Detail)
	}

	return nil
}

func (ns *NamesiloApi) AddDNSRecord(domain, domainType, host, value string, ttl int) error {
	reqValues := url.Values{}

	reqValues.Add("rrtype", domainType)
	reqValues.Add("rrhost", host)
	reqValues.Add("rrvalue", domain)
	reqValues.Add("rrttl", strconv.Itoa(ttl))
	reqValues.Add("rrdistance", "0")

	reqUrl, err := ns.apiActionWithValues("dnsAddRecord", &reqValues)
	if err != nil {
		return err
	}

	var darr DNSAddRecordsResponse
	if err := request(reqUrl, &darr); err != nil {
		return err
	} else if darr.Reply.Detail != "success" {
		return fmt.Errorf("namesilo domain add failed with: %s", darr.Reply.Detail)
	}

	return nil
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
	newValues.Add("domain", ns.domain)

	reqUrl.RawQuery = newValues.Encode()

	return reqUrl, nil
}

func request(url_ *url.URL, responseBody interface{}) error {
	response, err := http.Get(url_.String())
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return xml.Unmarshal(body, responseBody)
}
