package nsdns

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestNewDnsManager(t *testing.T) {
	dm, err := NewDnsManager("a", "b")
	assert.NoError(t, err)
	assert.Equal(t, "a", dm.BareDomainName)
	assert.Equal(t, "b", dm.TargetIngressClass)

	dm, err = NewDnsManager("", "b")
	assert.Nil(t, dm)
	assert.Equal(t, "must provide a domain name to target DNS record updates", err.Error())

	dm, err = NewDnsManager("a", "")
	assert.Nil(t, dm)
	assert.Equal(t, "must provide an ingress class to generate DNS records", err.Error())
}
