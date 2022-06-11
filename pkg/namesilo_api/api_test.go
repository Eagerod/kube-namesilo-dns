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

	api := NewNamesiloApiWithServer("api-key", server.URL)
	rr, err := api.ListDNSRecords("example.com")
	assert.NoError(t, err)

	assert.Equal(t, expectedCalls, calls)

	assert.Equal(t, 1, len(rr))
}

func TestUpdateDNSRecord(t *testing.T) {
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

	api := NewNamesiloApiWithServer("api-key", server.URL)
	err := api.UpdateDNSRecord("example.com", "sub", "abc123", "192.168.1.1", 1234)
	assert.NoError(t, err)

	assert.Equal(t, expectedCalls, calls)
}

func TestAddDNSRecord(t *testing.T) {
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
		assert.Equal(t, []string{"A"}, q["rrtype"])
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

	api := NewNamesiloApiWithServer("api-key", server.URL)
	err := api.AddDNSRecord("example.com", "A", "sub", "192.168.1.1", 1234)
	assert.NoError(t, err)

	assert.Equal(t, expectedCalls, calls)
}
