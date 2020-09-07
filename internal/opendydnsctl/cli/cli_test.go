package cli

import (
	"github.com/creekorful/open-dydns/internal/opendydnsctl/config"
	"github.com/creekorful/open-dydns/pkg/proto"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"testing"
)

func TestCli_Authenticate_InvalidRequest(t *testing.T) {
	c := cli{}

	if _, err := c.Authenticate(proto.CredentialsDto{}); err != ErrBadRequest {
		t.Error("Authenticate() should return ErrBadRequest")
	}
	if _, err := c.Authenticate(proto.CredentialsDto{Email: "test@example.org"}); err != ErrBadRequest {
		t.Error("Authenticate() should return ErrBadRequest")
	}
}

func TestCli_Authenticate_AlreadyLoggedIn(t *testing.T) {
	c := cli{
		conf: config.Config{
			Token: "tset",
		},
	}

	if _, err := c.Authenticate(proto.CredentialsDto{Email: "email", Password: "password"}); err != ErrAlreadyLoggedIn {
		t.Errorf("Authenticate() should have returned ErrAlreadyLoggedIn")
	}
}

func TestCli_Authenticate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	clientMock := proto.NewMockAPIContract(mockCtrl)
	configMock := config.NewMockProvider(mockCtrl)

	c := cli{
		logger:       &l,
		apiClient:    clientMock,
		confProvider: configMock,
	}

	clientMock.EXPECT().
		Authenticate(proto.CredentialsDto{Email: "root", Password: "toor"}).
		Return(proto.TokenDto{Token: "test-token"}, nil)
	configMock.EXPECT().Save(config.Config{Token: "test-token"})

	tok, err := c.Authenticate(proto.CredentialsDto{Email: "root", Password: "toor"})
	if err != nil {
		t.Error(err)
	}

	if tok.Token != "test-token" {
		t.Error("invalid token returned")
	}
}

func TestCli_GetAliases(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	clientMock := proto.NewMockAPIContract(mockCtrl)

	c := cli{
		logger:    &l,
		apiClient: clientMock,
		conf: config.Config{
			Aliases: map[string]config.AliasConfig{
				"creekorful.fr": {Synchronize: true},
			},
		},
		tok: proto.TokenDto{Token: "test-token"},
	}

	clientMock.EXPECT().GetAliases(c.tok).Return([]proto.AliasDto{
		{Domain: "creekorful.fr", Value: "127.0.0.1"},
		{Domain: "example.org", Value: "127.0.0.1"},
	}, nil)

	aliases, err := c.GetAliases()
	if err != nil {
		t.Error(err)
	}

	if len(aliases) != 2 {
		t.Error("wrong number of aliases returned")
	}

	for _, alias := range aliases {
		if alias.Domain == "creekorful.fr" && !alias.Synchronize {
			t.Error("alias creekorful.fr should have Synchronize = true")
		}
	}
}

func TestCli_RegisterAlias_InvalidRequest(t *testing.T) {
	c := cli{}

	if _, err := c.RegisterAlias(proto.AliasDto{}); err != ErrBadRequest {
		t.Error("RegisterAlias() should return ErrBadRequest")
	}
	if _, err := c.RegisterAlias(proto.AliasDto{Domain: "example.org"}); err != ErrBadRequest {
		t.Error("RegisterAlias() should return ErrBadRequest")
	}
}

func TestCli_RegisterAlias_AliasTaken(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	clientMock := proto.NewMockAPIContract(mockCtrl)

	c := cli{
		logger:    &l,
		apiClient: clientMock,
		tok:       proto.TokenDto{Token: "test-token"},
	}

	clientMock.EXPECT().
		RegisterAlias(c.tok, proto.AliasDto{Domain: "foo.bar.baz", Value: "127.0.0.1"}).
		Return(proto.AliasDto{}, proto.ErrAliasTaken)

	_, err := c.RegisterAlias(proto.AliasDto{Domain: "foo.bar.baz", Value: "127.0.0.1"})
	if err != proto.ErrAliasTaken {
		t.Error("RegisterAlias() should have returned proto.ErrAliasTaken")
	}
}

func TestCli_RegisterAlias(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	clientMock := proto.NewMockAPIContract(mockCtrl)

	c := cli{
		logger:    &l,
		apiClient: clientMock,
		tok:       proto.TokenDto{Token: "test-token"},
	}

	clientMock.EXPECT().
		RegisterAlias(c.tok, proto.AliasDto{Domain: "foo.bar.baz", Value: "127.0.0.1"}).
		Return(proto.AliasDto{Domain: "foo.bar.baz", Value: "127.0.0.1"}, nil)

	al, err := c.RegisterAlias(proto.AliasDto{Domain: "foo.bar.baz", Value: "127.0.0.1"})
	if err != nil {
		t.Error(err)
	}

	if al.Domain != "foo.bar.baz" || al.Value != "127.0.0.1" {
		t.Error("wrong alias returned")
	}
}

func TestCli_UpdateAlias_InvalidRequest(t *testing.T) {
	c := cli{}

	if _, err := c.UpdateAlias(proto.AliasDto{}); err != ErrBadRequest {
		t.Error("UpdateAlias() should return ErrBadRequest")
	}
	if _, err := c.UpdateAlias(proto.AliasDto{Domain: "example.org"}); err != ErrBadRequest {
		t.Error("UpdateAlias() should return ErrBadRequest")
	}
}

func TestCli_UpdateAlias_AliasNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	clientMock := proto.NewMockAPIContract(mockCtrl)

	c := cli{
		logger:    &l,
		apiClient: clientMock,
		tok:       proto.TokenDto{Token: "test-token"},
	}

	clientMock.EXPECT().
		UpdateAlias(c.tok, proto.AliasDto{Domain: "foo.bar.baz", Value: "127.0.0.1"}).
		Return(proto.AliasDto{}, proto.ErrAliasNotFound)

	_, err := c.UpdateAlias(proto.AliasDto{Domain: "foo.bar.baz", Value: "127.0.0.1"})
	if err != proto.ErrAliasNotFound {
		t.Error("UpdateAlias() should have returned proto.ErrAliasNotFound")
	}
}

func TestCli_UpdateAlias(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	clientMock := proto.NewMockAPIContract(mockCtrl)

	c := cli{
		logger:    &l,
		apiClient: clientMock,
		tok:       proto.TokenDto{Token: "test-token"},
	}

	clientMock.EXPECT().
		UpdateAlias(c.tok, proto.AliasDto{Domain: "foo.bar.baz", Value: "127.0.0.1"}).
		Return(proto.AliasDto{Domain: "foo.bar.baz", Value: "127.0.0.1"}, nil)

	al, err := c.UpdateAlias(proto.AliasDto{Domain: "foo.bar.baz", Value: "127.0.0.1"})
	if err != nil {
		t.Error(err)
	}

	if al.Domain != "foo.bar.baz" || al.Value != "127.0.0.1" {
		t.Error("wrong alias returned")
	}
}

func TestCli_DeleteAlias_AliasNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	clientMock := proto.NewMockAPIContract(mockCtrl)

	c := cli{
		logger:    &l,
		apiClient: clientMock,
		tok:       proto.TokenDto{Token: "test-token"},
	}

	clientMock.EXPECT().
		DeleteAlias(c.tok, "foo.bar.baz").
		Return(proto.ErrAliasNotFound)

	if err := c.DeleteAlias("foo.bar.baz"); err != proto.ErrAliasNotFound {
		t.Error("DeleteAlias() should have failed")
	}
}

func TestCli_DeleteAlias(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	clientMock := proto.NewMockAPIContract(mockCtrl)

	c := cli{
		logger:    &l,
		apiClient: clientMock,
		tok:       proto.TokenDto{Token: "test-token"},
	}

	clientMock.EXPECT().
		DeleteAlias(c.tok, "foo.bar.baz").
		Return(nil)

	if err := c.DeleteAlias("foo.bar.baz"); err != nil {
		t.Error("DeleteAlias() should not have failed")
	}
}

func TestCli_GetDomains(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	clientMock := proto.NewMockAPIContract(mockCtrl)

	c := cli{
		logger:    &l,
		apiClient: clientMock,
		tok:       proto.TokenDto{Token: "test-token"},
	}

	clientMock.EXPECT().
		GetDomains(c.tok).
		Return([]proto.DomainDto{{Domain: "creekorful.fr"}, {Domain: "example.org"}}, nil)

	domains, err := c.GetDomains()
	if err != nil {
		t.Error(err)
	}

	if len(domains) != 2 {
		t.Error("wrong number of domains returned")
	}
}

func TestCli_Synchronize(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	clientMock := proto.NewMockAPIContract(mockCtrl)

	c := cli{
		logger:    &l,
		apiClient: clientMock,
		tok:       proto.TokenDto{Token: "test-token"},
		conf: config.Config{
			Aliases: map[string]config.AliasConfig{
				"foo.bar.baz":        {Synchronize: false},
				"foo.example.org":    {Synchronize: true},
				"local.example.org":  {Synchronize: true},
				"dummy.notexist.org": {Synchronize: true},
			},
		},
	}

	clientMock.EXPECT().
		UpdateAlias(c.tok, proto.AliasDto{Domain: "local.example.org", Value: "127.0.0.1"}).
		Return(proto.AliasDto{Domain: "local.example.org", Value: "127.0.0.1"}, nil)
	clientMock.EXPECT().
		UpdateAlias(c.tok, proto.AliasDto{Domain: "foo.example.org", Value: "127.0.0.1"}).
		Return(proto.AliasDto{Domain: "foo.example.org", Value: "127.0.0.1"}, nil)
	clientMock.EXPECT().
		UpdateAlias(c.tok, proto.AliasDto{Domain: "dummy.notexist.org", Value: "127.0.0.1"}).
		Return(proto.AliasDto{}, proto.ErrAliasNotFound)

	if err := c.Synchronize("127.0.0.1"); err != nil {
		t.Error(err)
	}
}

func TestCli_SetSynchronize(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	confProvider := config.NewMockProvider(mockCtrl)

	c := cli{
		logger:       &l,
		confProvider: confProvider,
		conf: config.Config{
			Aliases: map[string]config.AliasConfig{
				"foo.bar.baz":     {Synchronize: false},
				"foo.example.org": {Synchronize: true},
			},
		},
	}

	confProvider.EXPECT().Save(config.Config{
		Aliases: map[string]config.AliasConfig{
			"foo.bar.baz":     {Synchronize: true},
			"foo.example.org": {Synchronize: true},
		},
	})

	if err := c.SetSynchronize("foo.bar.baz", true); err != nil {
		t.Error(err)
	}

	if !c.conf.Aliases["foo.bar.baz"].Synchronize {
		t.Error("alias foo.bar.baz is not updated")
	}

	confProvider.EXPECT().Save(config.Config{
		Aliases: map[string]config.AliasConfig{
			"foo.bar.baz":     {Synchronize: true},
			"foo.example.org": {Synchronize: false},
		},
	})

	if err := c.SetSynchronize("foo.example.org", false); err != nil {
		t.Error(err)
	}

	if c.conf.Aliases["foo.example.org"].Synchronize {
		t.Error("alias foo.example.org is not updated")
	}
}
