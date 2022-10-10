package namesilo_api

import (
	"encoding/xml"
)

type ResourceRecord struct {
	XMLName  xml.Name `xml:"resource_record"`
	RecordId string   `xml:"record_id"`
	Type     string   `xml:"type"`
	Host     string   `xml:"host"`
	Value    string   `xml:"value"`
	TTL      int      `xml:"ttl"`
	Distance int      `xml:"distance"`
}

func (r ResourceRecord) Equals(other interface{}) bool {
	if r == other {
		return true
	}

	otherRR, ok := other.(ResourceRecord)
	if !ok {
		return false
	}

	matchIds := true
	if r.RecordId != "" && otherRR.RecordId != "" {
		matchIds = r.RecordId == otherRR.RecordId
	}
	return matchIds &&
		r.Type == otherRR.Type &&
		r.Host == otherRR.Host &&
		r.Value == otherRR.Value &&
		r.TTL == otherRR.TTL &&
		r.Distance == otherRR.Distance
}
