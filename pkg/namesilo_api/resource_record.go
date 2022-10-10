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
	if other == nil {
		return false
	}

	otherRRp, ok := other.(*ResourceRecord)
	if ok {
		return r.EqualsRecord(*otherRRp)
	}

	otherRR, ok := other.(ResourceRecord)
	if !ok {
		return false
	}

	return r.EqualsRecord(otherRR)
}

func (r ResourceRecord) EqualsRecord(other ResourceRecord) bool {
	matchIds := true
	if r.RecordId != "" && other.RecordId != "" {
		matchIds = r.RecordId == other.RecordId
	}
	return matchIds &&
		r.Type == other.Type &&
		r.Host == other.Host &&
		r.Value == other.Value &&
		r.TTL == other.TTL &&
		r.Distance == other.Distance
}
