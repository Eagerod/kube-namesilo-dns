package namesilo_api

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestListDNSRecord(t *testing.T) {
	expectedCalls := 1
	calls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dnsListRecords", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, []string{"1"}, q["version"])
		assert.Equal(t, []string{"xml"}, q["type"])
		assert.Equal(t, []string{"api-key"}, q["key"])
		assert.Equal(t, []string{"example.com"}, q["domain"])

		var response ListDNSRecordsResponse
		response.Reply.ResourceRecords = append(response.Reply.ResourceRecords, ResourceRecord{})
		response.Reply.Detail = "success"
		body, err := xml.Marshal(response)
		assert.NoError(t, err)
		w.Write(body)

		calls += 1
	}))
	defer server.Close()

	api := NewNamesiloApiWithServer("example.com", "api-key", server.URL)
	rr, err := api.ListDNSRecords()
	assert.NoError(t, err)

	assert.Equal(t, expectedCalls, calls)

	assert.Equal(t, 1, len(rr))
}

func TestUpdateDNSARecord(t *testing.T) {
	expectedCalls := 1
	calls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dnsUpdateRecord", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, []string{"1"}, q["version"])
		assert.Equal(t, []string{"xml"}, q["type"])
		assert.Equal(t, []string{"api-key"}, q["key"])
		assert.Equal(t, []string{"example.com"}, q["domain"])
		assert.Equal(t, []string{"abc123"}, q["rrid"])
		assert.Equal(t, []string{"sub"}, q["rrhost"])
		assert.Equal(t, []string{"192.168.1.1"}, q["rrvalue"])
		assert.Equal(t, []string{"1234"}, q["rrttl"])

		var response DNSUpdateRecordsResponse
		response.Reply.Detail = "success"
		body, err := xml.Marshal(response)
		assert.NoError(t, err)
		w.Write(body)

		calls += 1
	}))
	defer server.Close()

	record := ResourceRecord{
		RecordId: "abc123",
		Type: "A",
		Host: "sub.example.com",
		Value: "192.168.1.1",
		TTL: 1234,
		Distance :0,
	}

	api := NewNamesiloApiWithServer("example.com", "api-key", server.URL)
	err := api.UpdateDNSRecord(record)
	assert.NoError(t, err)

	assert.Equal(t, expectedCalls, calls)
}


func TestUpdateDNSCnameRecord(t *testing.T) {
	expectedCalls := 1
	calls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dnsUpdateRecord", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, []string{"1"}, q["version"])
		assert.Equal(t, []string{"xml"}, q["type"])
		assert.Equal(t, []string{"api-key"}, q["key"])
		assert.Equal(t, []string{"example.com"}, q["domain"])
		assert.Equal(t, []string{"abc123"}, q["rrid"])
		assert.Equal(t, []string{"sub"}, q["rrhost"])
		assert.Equal(t, []string{"example.com"}, q["rrvalue"])
		assert.Equal(t, []string{"1234"}, q["rrttl"])

		var response DNSUpdateRecordsResponse
		response.Reply.Detail = "success"
		body, err := xml.Marshal(response)
		assert.NoError(t, err)
		w.Write(body)

		calls += 1
	}))
	defer server.Close()

	record := ResourceRecord{
		RecordId: "abc123",
		Type: "CNAME",
		Host: "sub.example.com",
		Value: "example.com",
		TTL: 1234,
		Distance :0,
	}

	api := NewNamesiloApiWithServer("example.com", "api-key", server.URL)
	err := api.UpdateDNSRecord(record)
	assert.NoError(t, err)

	assert.Equal(t, expectedCalls, calls)
}

func TestUpdateDNSRecordFailsWithNoRRID(t *testing.T) {
	expectedCalls := 0
	calls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls += 1
	}))
	defer server.Close()

	record := ResourceRecord{
		Type: "CNAME",
		Host: "sub.example.com",
		Value: "example.com",
		TTL: 1234,
		Distance :0,
	}

	api := NewNamesiloApiWithServer("example.com", "api-key", server.URL)
	err := api.UpdateDNSRecord(record)
	assert.Equal(t, err.Error(), "cannot update DNS record without record id")

	assert.Equal(t, expectedCalls, calls)
}

func TestAddDNSRecordARecord(t *testing.T) {
	expectedCalls := 1
	calls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dnsAddRecord", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, []string{"1"}, q["version"])
		assert.Equal(t, []string{"xml"}, q["type"])
		assert.Equal(t, []string{"api-key"}, q["key"])
		assert.Equal(t, []string{"example.com"}, q["domain"])
		assert.Equal(t, []string{""}, q["rrhost"])
		assert.Equal(t, []string{"A"}, q["rrtype"])
		assert.Equal(t, []string{"192.168.1.1"}, q["rrvalue"])
		assert.Equal(t, []string{"1234"}, q["rrttl"])
		assert.Equal(t, []string{"0"}, q["rrdistance"])

		var response DNSUpdateRecordsResponse
		response.Reply.Detail = "success"
		body, err := xml.Marshal(response)
		assert.NoError(t, err)
		w.Write(body)

		calls += 1
	}))
	defer server.Close()

	record := ResourceRecord{
		Type: "A",
		Host: "example.com",
		Value: "192.168.1.1",
		TTL: 1234,
		Distance :0,
	}

	api := NewNamesiloApiWithServer("example.com", "api-key", server.URL)
	err := api.AddDNSRecord(record)
	assert.NoError(t, err)

	assert.Equal(t, expectedCalls, calls)
}

func TestAddDNSRecordCnameRecord(t *testing.T) {
	expectedCalls := 1
	calls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dnsAddRecord", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, []string{"1"}, q["version"])
		assert.Equal(t, []string{"xml"}, q["type"])
		assert.Equal(t, []string{"api-key"}, q["key"])
		assert.Equal(t, []string{"example.com"}, q["domain"])
		assert.Equal(t, []string{"sub"}, q["rrhost"])
		assert.Equal(t, []string{"CNAME"}, q["rrtype"])
		assert.Equal(t, []string{"example.com"}, q["rrvalue"])
		assert.Equal(t, []string{"1234"}, q["rrttl"])
		assert.Equal(t, []string{"0"}, q["rrdistance"])

		var response DNSUpdateRecordsResponse
		response.Reply.Detail = "success"
		body, err := xml.Marshal(response)
		assert.NoError(t, err)
		w.Write(body)

		calls += 1
	}))
	defer server.Close()

	record := ResourceRecord{
		Type: "CNAME",
		Host: "sub.example.com",
		Value: "example.com",
		TTL: 1234,
		Distance :0,
	}

	api := NewNamesiloApiWithServer("example.com", "api-key", server.URL)
	err := api.AddDNSRecord(record)
	assert.NoError(t, err)

	assert.Equal(t, expectedCalls, calls)
}

func TestDeleteDNSRecord(t *testing.T) {
	expectedCalls := 1
	calls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dnsDeleteRecord", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, []string{"1"}, q["version"])
		assert.Equal(t, []string{"xml"}, q["type"])
		assert.Equal(t, []string{"api-key"}, q["key"])
		assert.Equal(t, []string{"example.com"}, q["domain"])
		assert.Equal(t, []string{"abc123"}, q["rrid"])

		var response DNSUpdateRecordsResponse
		response.Reply.Detail = "success"
		body, err := xml.Marshal(response)
		assert.NoError(t, err)
		w.Write(body)

		calls += 1
	}))
	defer server.Close()

	record := ResourceRecord{
		RecordId: "abc123",
		Type: "CNAME",
		Host: "sub.example.com",
		Value: "example.com",
		TTL: 1234,
		Distance :0,
	}

	api := NewNamesiloApiWithServer("example.com", "api-key", server.URL)
	err := api.DeleteDNSRecord(record)
	assert.NoError(t, err)

	assert.Equal(t, expectedCalls, calls)
}


func TestDeleteDNSRecordFailsWithoutId(t *testing.T) {
	expectedCalls := 0
	calls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls += 1
	}))
	defer server.Close()

	record := ResourceRecord{
		Type: "CNAME",
		Host: "sub.example.com",
		Value: "example.com",
		TTL: 1234,
		Distance :0,
	}

	api := NewNamesiloApiWithServer("example.com", "api-key", server.URL)
	err := api.DeleteDNSRecord(record)
	assert.Equal(t, err.Error(), "cannot delete DNS record without ID")

	assert.Equal(t, expectedCalls, calls)
}
