package daemon

import (
	"github.com/creekorful/open-dydns/internal/opendydnsd/database"
	"github.com/creekorful/open-dydns/internal/proto"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	d := daemon{}

	pass, err := d.hashPassword("test")
	if err != nil {
		t.Error("unable to hash password")
	}

	if !d.validatePassword(pass, "test") {
		t.Error("password should be valid")
	}
}

func TestNewAliasDto(t *testing.T) {
	alias := newAliasDto(database.Alias{
		Domain: "domain",
		Value:  "value",
	})

	if alias.Domain != "domain" {
		t.FailNow()
	}
	if alias.Value != "value" {
		t.FailNow()
	}
}

func TestNewAlias(t *testing.T) {
	alias := newAlias(proto.AliasDto{
		Domain: "domain",
		Value:  "value",
	})

	if alias.Domain != "domain" {
		t.FailNow()
	}
	if alias.Value != "value" {
		t.FailNow()
	}
}
