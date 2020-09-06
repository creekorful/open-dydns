package dns

import "fmt"

//go:generate mockgen -source provisioner.go -destination=./provisioner_mock.go -package=dns

// Provisioner represent a DNS provisioner
// i.e used to abstract different DNS provisioner API solutions
type Provisioner interface {
	AddRecord(host, domain, value string) error
	UpdateRecord(host, domain, value string) error
	DeleteRecord(host, domain string) error
}

// Provider is the abstraction used to resolve a Provisioner
// based on his name etc. This ease unit testing
type Provider interface {
	GetProvisioner(name string, config map[string]string) (Provisioner, error)
}

type provider struct {
}

// NewProvider return the default Provider implementation
func NewProvider() Provider {
	return &provider{}
}

// GetProvisioner return the appropriate Provisioner based on his name
func (p *provider) GetProvisioner(name string, config map[string]string) (Provisioner, error) {
	switch name {
	case ovhProvisionerName:
		return newOVHProvisioner(config)
	default:
		return nil, fmt.Errorf("no provisioner named %s found", name)
	}
}

func getConfigOrFail(config map[string]string, name string) (string, error) {
	val := ""
	if v, exist := config[name]; exist {
		val = v
	} else {
		return "", fmt.Errorf("missing config `%s`", name)
	}
	return val, nil
}
