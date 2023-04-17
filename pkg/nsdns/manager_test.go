package nsdns

import (
	"os"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apinetworkingv1 "k8s.io/api/networking/v1"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/namesilo_api"
)

type MockNamesiloApi struct{
	mock.Mock
}

func (nsapi *MockNamesiloApi) ListDNSRecords() ([]namesilo_api.ResourceRecord, error) {
	args := nsapi.Called()
	return args.Get(0).([]namesilo_api.ResourceRecord), args.Error(1)
}

func (nsapi *MockNamesiloApi) UpdateDNSRecord(rr namesilo_api.ResourceRecord) error {
	args := nsapi.Called(rr)
	return args.Error(0)
}

func (nsapi *MockNamesiloApi) AddDNSRecord(rr namesilo_api.ResourceRecord) error {
	args := nsapi.Called(rr)
	return args.Error(0)
}

func (nsapi *MockNamesiloApi) DeleteDNSRecord(rr namesilo_api.ResourceRecord) error {
	args := nsapi.Called(rr)
	return args.Error(0)
}

func TestNewDnsManager(t *testing.T) {
	ov := os.Getenv("NAMESILO_API_KEY")
	os.Setenv("NAMESILO_API_KEY", "a")

	dm, err := NewDnsManager("a", "b")
	assert.NoError(t, err)
	assert.Equal(t, "a", dm.BareDomainName)
	assert.Equal(t, "b", dm.TargetIngressClass)

	os.Setenv("NAMESILO_API_KEY", "")
	_, err = NewDnsManager("a", "b")
	assert.Equal(t, "failed to find NAMESILO_API_KEY in environment; cannot proceed", err.Error())

	os.Setenv("NAMESILO_API_KEY", ov)
}

func TestNewDnsManagerWithApiKey(t *testing.T) {
	dm, err := NewDnsManagerWithApiKey("a", "b", "c")
	assert.NoError(t, err)
	assert.Equal(t, "a", dm.BareDomainName)
	assert.Equal(t, "b", dm.TargetIngressClass)

	dm, err = NewDnsManagerWithApiKey("", "b", "c")
	assert.Nil(t, dm)
	assert.Equal(t, "must provide a domain name to target DNS record updates", err.Error())

	dm, err = NewDnsManagerWithApiKey("a", "", "c")
	assert.Nil(t, dm)
	assert.Equal(t, "must provide an ingress class to generate DNS records", err.Error())
}

func TestShouldProcessIngress(t *testing.T) {
	dm, err := NewDnsManagerWithApiKey("a", "b", "c")
	assert.NoError(t, err)

	ingress := apinetworkingv1.Ingress{}
	ingress.Annotations = map[string]string{}
	ingress.Annotations["kubernetes.io/ingress.class"] = dm.TargetIngressClass

	assert.True(t, dm.ShouldProcessIngress(&ingress))

	ingress.Annotations["kubernetes.io/ingress.class"] = dm.TargetIngressClass + "not"

	assert.False(t, dm.ShouldProcessIngress(&ingress))
}

func TestHandleIngressExists(t *testing.T) {
	dm, err := NewDnsManagerWithApiKey("example.com", "b", "c")
	assert.NoError(t, err)

	nsapi := MockNamesiloApi{}
	dm.Api = &nsapi
	dm.cache.CurrentIpAddress = "1.1.1.1"

	ingress := apinetworkingv1.Ingress{}
	ingress.Annotations = map[string]string{}
	ingress.Annotations["kubernetes.io/ingress.class"] = dm.TargetIngressClass
	ingress.Spec.Rules = append(ingress.Spec.Rules, apinetworkingv1.IngressRule{})
	ingress.Spec.Rules[0].Host = "example.com"

	expectedArg := namesilo_api.ResourceRecord{
		Type:     "A",
		Host:     "example.com",
		Value:    "1.1.1.1",
		TTL:      7207,
		Distance: 0,
	}

	m := nsapi.On("AddDNSRecord", expectedArg).Return(nil)

	// Create
	err = dm.HandleIngressExists(&ingress)
	assert.NoError(t, err)

	nsapi.AssertExpectations(t)
	m.Unset()

	rr := namesilo_api.ResourceRecord{
		RecordId: "1234",
		Type:     "A",
		Host:     "example.com",
		Value:    "1.1.1.1",
		TTL:      7207,
		Distance: 0,
	}
	dm.cache.CurrentRecords = append(dm.cache.CurrentRecords, rr)

	// No op
	err = dm.HandleIngressExists(&ingress)
	assert.NoError(t, err)

	nsapi.AssertExpectations(t)

	dm.cache.CurrentRecords = append(dm.cache.CurrentRecords, rr)
	dm.cache.CurrentRecords[0].Value = "1.1.1.2"
	expectedArg.RecordId = "1234"

	// Update
	m = nsapi.On("UpdateDNSRecord", expectedArg).Return(nil)
	err = dm.HandleIngressExists(&ingress)
	assert.NoError(t, err)

	nsapi.AssertExpectations(t)
	m.Unset()

	// Wrong ingress class
	ingress.Annotations["kubernetes.io/ingress.class"] = dm.TargetIngressClass + "not"

	err = dm.HandleIngressExists(&ingress)
	assert.NoError(t, err)

	nsapi.AssertExpectations(t)
}

func TestHandleIngressDeleted(t *testing.T) {
	dm, err := NewDnsManagerWithApiKey("example.com", "b", "c")
	assert.NoError(t, err)

	nsapi := MockNamesiloApi{}
	dm.Api = &nsapi

	ingress := apinetworkingv1.Ingress{}
	ingress.Annotations = map[string]string{}
	ingress.Annotations["kubernetes.io/ingress.class"] = dm.TargetIngressClass
	ingress.Spec.Rules = append(ingress.Spec.Rules, apinetworkingv1.IngressRule{})
	ingress.Spec.Rules[0].Host = "example.com"

	err = dm.HandleIngressDeleted(&ingress)
	assert.Equal(t, "failed to find record: A:example.com", err.Error())
	nsapi.AssertExpectations(t)

	rr := namesilo_api.ResourceRecord{
		RecordId: "1234",
		Type:     "A",
		Host:     "example.com",
		Value:    "1.1.1.1",
		TTL:      7207,
		Distance: 0,
	}
	dm.cache.CurrentRecords = append(dm.cache.CurrentRecords, rr)
	m := nsapi.On("DeleteDNSRecord", rr).Return(nil)

	err = dm.HandleIngressDeleted(&ingress)
	assert.NoError(t, err)

	nsapi.AssertExpectations(t)
	m.Unset()

	ingress = apinetworkingv1.Ingress{}
	ingress.Annotations = map[string]string{}
	ingress.Annotations["kubernetes.io/ingress.class"] = dm.TargetIngressClass + "not"

	err = dm.HandleIngressDeleted(&ingress)
	assert.NoError(t, err)

	nsapi.AssertExpectations(t)
}
