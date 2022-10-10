package namesilo_api

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestEquals(t *testing.T) {
	rrWithID1 := ResourceRecord{
		RecordId: "abc123",
		Type:     "CNAME",
		Host:     "sub.example.com",
		Value:    "example.com",
		TTL:      1234,
		Distance: 0,
	}
	rrWithID2 := ResourceRecord{
		RecordId: "123abc",
		Type:     "CNAME",
		Host:     "sub.example.com",
		Value:    "example.com",
		TTL:      1234,
		Distance: 0,
	}
	rrNoID1 := ResourceRecord{
		Type:     "CNAME",
		Host:     "sub.example.com",
		Value:    "example.com",
		TTL:      1234,
		Distance: 0,
	}
	rrNoID2 := ResourceRecord{
		Type:     "CNAME",
		Host:     "subdomain.example.com",
		Value:    "example.com",
		TTL:      1234,
		Distance: 0,
	}
	var tests = []struct {
		name      string
		input     ResourceRecord
		compareTo interface{}
		equal     bool
	}{
		{"EqualWithIds", rrWithID1, rrWithID1, true},
		{"EqualNoIds", rrNoID1, rrNoID1, true},
		{"EqualRef", rrWithID1, &rrWithID1, true},
		{"EqualDifferentIds", rrWithID1, rrWithID2, false},
		{"EqualDifferentProps", rrNoID1, rrNoID2, false},
		{"EqualOneMissingId", rrWithID1, rrNoID1, true},
		{"EqualWithNil", rrWithID1, nil, false},
		{"EqualWithOtherType", rrWithID1, "Legit Resource Record", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			equal := tt.input.Equals(tt.compareTo)
			assert.Equal(t, tt.equal, equal)
		})
	}
}
