package dns

import "testing"

func TestNewOvhProvisioner(t *testing.T) {
	if _, err := newOVHProvisioner(map[string]string{}); err == nil {
		t.Error("newOVHProvisioner should have failed")
	}

	if _, err := newOVHProvisioner(map[string]string{
		"endpoint":     "ovh-eu",
		"app-key":      "test",
		"app-secret":   "test",
		"consumer-key": "test",
	}); err != nil {
		t.Error("newOVHProvisioner has failed")
	}
}
