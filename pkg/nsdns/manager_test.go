package nsdns

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
	apinetworkingv1 "k8s.io/api/networking/v1"
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

func TestShouldProcessIngress(t *testing.T) {
	dm, err := NewDnsManager("a", "b")
	assert.NoError(t, err)

	ingress := apinetworkingv1.Ingress{}
	ingress.Annotations = map[string]string{}
	ingress.Annotations["kubernetes.io/ingress.class"] = dm.TargetIngressClass

	assert.True(t, dm.ShouldProcessIngress(&ingress))

	ingress.Annotations["kubernetes.io/ingress.class"] = dm.TargetIngressClass + "not"

	assert.False(t, dm.ShouldProcessIngress(&ingress))
}
