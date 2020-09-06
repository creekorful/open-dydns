package dns

import (
	"github.com/ovh/go-ovh/ovh"
)

const ovhProvisionerName = "ovh"

type ovhProvisioner struct {
	client *ovh.Client
}

func newOVHProvisioner(config map[string]string) (Provisioner, error) {
	endpoint, err := getConfigOrFail(config, "endpoint")
	if err != nil {
		return nil, err
	}
	appKey, err := getConfigOrFail(config, "app-key")
	if err != nil {
		return nil, err
	}
	appSecret, err := getConfigOrFail(config, "app-secret")
	if err != nil {
		return nil, err
	}
	consumerKey, err := getConfigOrFail(config, "consumer-key")
	if err != nil {
		return nil, err
	}

	client, err := ovh.NewClient(endpoint, appKey, appSecret, consumerKey)
	if err != nil {
		return nil, err
	}

	return &ovhProvisioner{
		client: client,
	}, nil
}

func (o *ovhProvisioner) AddRecord(host, domain, value string) error {
	panic("implement me")
}

func (o *ovhProvisioner) UpdateRecord(host, domain, value string) error {
	panic("implement me")
}

func (o *ovhProvisioner) DeleteRecord(host, domain string) error {
	panic("implement me")
}
