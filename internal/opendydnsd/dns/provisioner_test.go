package dns

import "testing"

func TestGetConfigOrFail(t *testing.T) {
	_, err := getConfigOrFail(map[string]string{}, "test")
	if err == nil {
		t.Error()
	}

	val, err := getConfigOrFail(map[string]string{"hello": "world"}, "hello")
	if val != "world" {
		t.Error()
	}
}
