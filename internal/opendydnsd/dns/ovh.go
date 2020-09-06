package dns

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
)

const (
	ovhProvisionerName = "ovh"
	zoneEndpoint       = "/domain/zone"
)

type ovhRecord struct {
	ID        int64  `json:"id,omitempty"`
	ZoneName  string `json:"zoneName,omitempty"`
	FieldType string `json:"fieldType"`
	SubDomain string `json:"subDomain"`
	Target    string `json:"target"`
	TTL       int64  `json:"ttl"`
}

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
	// add the record
	if err := o.client.Post(fmt.Sprintf("%s/%s/record", zoneEndpoint, domain), &ovhRecord{
		FieldType: "A", // TODO AAA if ipv6
		SubDomain: host,
		Target:    value,
	}, nil); err != nil {
		return err
	}

	// refresh the zone to apply changes
	return o.refreshZone(domain)
}

func (o *ovhProvisioner) UpdateRecord(host, domain, value string) error {
	record, err := o.findRecord(host, domain)
	if err != nil {
		return err
	}

	// update target
	record.Target = value

	url := fmt.Sprintf("%s/%s/record/%d", zoneEndpoint, domain, record.ID)
	if err := o.client.Put(url, &record, nil); err != nil {
		return err
	}

	return o.refreshZone(domain)
}

func (o *ovhProvisioner) DeleteRecord(host, domain string) error {
	// find the record to delete
	record, err := o.findRecord(host, domain)
	if err != nil {
		return err
	}

	// delete the record if found
	if err := o.client.Delete(fmt.Sprintf("%s/%s/record/%d", zoneEndpoint, domain, record.ID), nil); err != nil {
		return err
	}

	return o.refreshZone(domain)
}

func (o *ovhProvisioner) refreshZone(domain string) error {
	return o.client.Post(fmt.Sprintf("%s/%s/refresh", zoneEndpoint, domain), nil, nil)
}

func (o *ovhProvisioner) findRecord(host, domain string) (ovhRecord, error) {
	var recordIds []int64

	// Search for the record
	url := fmt.Sprintf("%s/%s/record?fieldType=A&subDomain=%s", zoneEndpoint, domain, host) // TODO manage Ipv6
	if err := o.client.Get(url, &recordIds); err != nil {
		return ovhRecord{}, err
	}

	if len(recordIds) != 1 {
		return ovhRecord{}, fmt.Errorf("more or less than 1 record found")
	}

	// Query for record details
	var record ovhRecord
	if err := o.client.Get(fmt.Sprintf("%s/%s/record/%d", zoneEndpoint, domain, recordIds[0]), &record); err != nil {
		return ovhRecord{}, err
	}

	return record, nil
}
