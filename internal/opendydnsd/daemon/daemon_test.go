package daemon

import (
	"errors"
	"github.com/creekorful/open-dydns/internal/opendydnsd/config"
	"github.com/creekorful/open-dydns/internal/opendydnsd/database"
	"github.com/creekorful/open-dydns/internal/opendydnsd/dns"
	"github.com/creekorful/open-dydns/pkg/proto"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"io/ioutil"
	"testing"
)

// TODO test provisioning fails case
// TODO cleanup code?

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
		Domain: "bar.baz",
		Host:   "foo",
		Value:  "value",
	})

	if alias.Domain != "foo.bar.baz" {
		t.FailNow()
	}
	if alias.Value != "value" {
		t.FailNow()
	}
}

func TestNewAlias(t *testing.T) {
	alias := newAlias(proto.AliasDto{
		Domain: "foo.bar.baz",
		Value:  "value",
	})

	if alias.Domain != "bar.baz" {
		t.FailNow()
	}
	if alias.Host != "foo" {
		t.FailNow()
	}
	if alias.Value != "value" {
		t.FailNow()
	}
}

func TestIsAliasValid(t *testing.T) {
	if isAliasValid(proto.AliasDto{
		Domain: "foo",
		Value:  "127.0.0.1",
	}) {
		t.Error("isAliasValid() should have return false")
	}

	if !isAliasValid(proto.AliasDto{
		Domain: "foo.bar.baz",
		Value:  "127.0.0.1",
	}) {
		t.Error("isAliasValid() should have return true")
	}
}

func TestDaemon_CreateUser_InvalidRequest(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	if _, err := d.CreateUser(proto.CredentialsDto{Email: "test@gmail.com"}); err != ErrInvalidParameters {
		t.Errorf("CreateUser() should have returned ErrInvalidParameters")
	}
}

func TestDaemon_CreateUser_EmailTaken(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	dbMock.EXPECT().FindUser("lunamicard@gmail.com").Return(database.User{}, nil)

	if _, err := d.CreateUser(proto.CredentialsDto{Email: "lunamicard@gmail.com", Password: "test"}); err != ErrInvalidParameters {
		t.Error("CreateUser() should have returned ErrInvalidParameters")
	}
}

func TestDaemon_CreateUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	dbMock.EXPECT().
		FindUser("lunamicard@gmail.com").
		Return(database.User{}, gorm.ErrRecordNotFound)
	dbMock.EXPECT().
		CreateUser("lunamicard@gmail.com", gomock.Any()).
		Return(database.User{}, nil)
	dbMock.EXPECT().
		FindUser("lunamicard@gmail.com").
		Return(database.User{Password: "$2a$04$5eQwROjKESuWP2y.sAVsPeqhG48UXWw.htYp5G./JsRjWwUMOi7xC"}, nil)

	if _, err := d.CreateUser(proto.CredentialsDto{Email: "lunamicard@gmail.com", Password: "test"}); err != nil {
		t.Errorf("CreateUser() should not have failed: %s", err)
	}
}

func TestDaemon_Authenticate_InvalidRequest(t *testing.T) {
	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	d := daemon{
		logger: &logger,
	}

	_, err := d.Authenticate(proto.CredentialsDto{})
	if !errors.As(err, &ErrInvalidParameters) {
		t.Error("Authenticate() should have failed")
	}
}

func TestDaemon_Authenticate_NonExistingUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	dbMock.EXPECT().
		FindUser("lunamicard@gmail.com").
		Return(database.User{}, gorm.ErrRecordNotFound)

	_, err := d.Authenticate(proto.CredentialsDto{Email: "lunamicard@gmail.com", Password: "test"})
	if !errors.As(err, &ErrUserNotFound) {
		t.Error("Authenticate() should have returned ErrUserNotFound")
	}
}

func TestDaemon_Authenticate_InvalidPassword(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	pass, err := d.hashPassword("test")
	if err != nil {
		t.Error(err)
	}

	dbMock.EXPECT().
		FindUser("lunamicard@gmail.com").
		Return(database.User{Email: "lunamicard@gmail.com", Password: pass}, nil)

	_, err = d.Authenticate(proto.CredentialsDto{Email: "lunamicard@gmail.com", Password: "testa"})
	if !errors.As(err, &ErrUserNotFound) {
		t.Error("Authenticate() should have returned ErrUserNotFound")
	}
}

func TestDaemon_Authenticate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	pass, err := d.hashPassword("test")
	if err != nil {
		t.Error(err)
	}

	dbMock.EXPECT().
		FindUser("lunamicard@gmail.com").
		Return(database.User{
			Model:    gorm.Model{ID: 1},
			Email:    "lunamicard@gmail.com",
			Password: pass,
			Aliases:  nil,
		}, nil)

	u, err := d.Authenticate(proto.CredentialsDto{Email: "lunamicard@gmail.com", Password: "test"})
	if err != nil {
		t.Error(err)
	}

	if u.UserID != 1 {
		t.Error("wrong userID")
	}
}

func TestDaemon_GetAliases(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	dbMock.EXPECT().
		FindUserAliases(uint(1)).
		Return([]database.Alias{{Domain: "bar.baz", Host: "foo", Value: "8.8.8.8"}}, nil)

	aliases, err := d.GetAliases(proto.UserContext{UserID: 1})
	if err != nil {
		t.Error(err)
	}

	if len(aliases) != 1 {
		t.Error("wrong number of aliases")
	}

	alias := aliases[0]
	if alias.Domain != "foo.bar.baz" || alias.Value != "8.8.8.8" {
		t.Error("Wrong alias returned")
	}
}

func TestDaemon_RegisterAlias_InvalidRequest(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	_, err := d.RegisterAlias(proto.UserContext{UserID: 1}, proto.AliasDto{})
	if !errors.As(err, &ErrInvalidParameters) {
		t.Error("RegisterAlias() should have returned ErrInvalidParameters")
	}

	_, err = d.RegisterAlias(proto.UserContext{UserID: 1}, proto.AliasDto{Domain: "test", Value: "8.8.8.8"})
	if !errors.As(err, &ErrInvalidParameters) {
		t.Error("RegisterAlias() should have returned ErrInvalidParameters")
	}
}

func TestDaemon_RegisterAlias_AliasTaken(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	dbMock.EXPECT().FindAlias("www", "creekorful.de").Return(database.Alias{
		Domain: "creekorful.de",
		Host:   "www",
		UserID: 12,
	}, nil)

	_, err := d.RegisterAlias(proto.UserContext{UserID: 1}, proto.AliasDto{
		Domain: "www.creekorful.de", Value: "127.0.0.1",
	})

	if !errors.As(err, &ErrAliasTaken) {
		t.Error("RegisterAlias() should have returned ErrAliasTaken")
	}
}

func TestDaemon_RegisterAlias_AliasAlreadyExist(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	dbMock.EXPECT().FindAlias("www", "creekorful.de").Return(database.Alias{
		Domain: "creekorful.de",
		Host:   "www",
		UserID: 1,
	}, nil)

	_, err := d.RegisterAlias(proto.UserContext{UserID: 1}, proto.AliasDto{
		Domain: "www.creekorful.de", Value: "127.0.0.1",
	})

	if !errors.As(err, &ErrAliasAlreadyExist) {
		t.Error("RegisterAlias() should have returned ErrAliasAlreadyExist")
	}
}

func TestDaemon_RegisterAlias(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)
	provisionerMock := dns.NewMockProvisioner(mockCtrl)
	providerMock := dns.NewMockProvider(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
		config: config.DaemonConfig{
			DNSProvisioners: []config.DNSProvisionerConfig{
				{
					Name:    "dummy",
					Config:  map[string]string{},
					Domains: []string{"creekorful.de"},
				},
			},
		},
		dnsProvider: providerMock,
	}

	dbMock.EXPECT().
		FindAlias("www", "creekorful.de").
		Return(database.Alias{}, gorm.ErrRecordNotFound)

	providerMock.EXPECT().GetProvisioner("dummy", map[string]string{}).Return(provisionerMock, nil)
	provisionerMock.EXPECT().AddRecord("www", "creekorful.de", "127.0.0.1").Return(nil)

	dbMock.EXPECT().
		CreateAlias(database.Alias{Domain: "creekorful.de", Host: "www", Value: "127.0.0.1"}, uint(1)).
		Return(database.Alias{
			Model:  gorm.Model{ID: 12},
			Domain: "creekorful.de",
			Host:   "www",
			Value:  "127.0.0.1",
			UserID: 1,
		}, nil)

	r, err := d.RegisterAlias(proto.UserContext{UserID: 1}, proto.AliasDto{
		Domain: "www.creekorful.de", Value: "127.0.0.1",
	})

	if err != nil {
		t.Error(err)
	}

	if r.Domain != "www.creekorful.de" || r.Value != "127.0.0.1" {
		t.Error("Wrong alias created")
	}
}

func TestDaemon_UpdateAlias_InvalidAlias(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	_, err := d.UpdateAlias(proto.UserContext{UserID: 1}, proto.AliasDto{Domain: "bar.baz", Value: "127.0.0.1"})
	if err != ErrInvalidParameters {
		t.Error("UpdateAlias() should have returned ErrInvalidParameters")
	}
}

func TestDaemon_UpdateAlias_AliasDoesNotExist(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	dbMock.EXPECT().
		FindAlias("foo", "bar.baz").
		Return(database.Alias{}, gorm.ErrRecordNotFound)

	_, err := d.UpdateAlias(proto.UserContext{UserID: 1}, proto.AliasDto{Domain: "foo.bar.baz", Value: "127.0.0.1"})
	if err != ErrAliasNotFound {
		t.Error("UpdateAlias() should have returned ErrAliasNotFound")
	}
}

func TestDaemon_UpdateAlias_AliasNotOwned(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
	}

	dbMock.EXPECT().
		FindAlias("foo", "bar.baz").
		Return(database.Alias{
			UserID: 12,
		}, nil)

	_, err := d.UpdateAlias(proto.UserContext{UserID: 1}, proto.AliasDto{Domain: "foo.bar.baz", Value: "127.0.0.1"})
	if err != ErrAliasNotFound {
		t.Error("UpdateAlias() should have returned ErrAliasNotFound")
	}
}

func TestDaemon_UpdateAlias(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)
	provisionerMock := dns.NewMockProvisioner(mockCtrl)
	providerMock := dns.NewMockProvider(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
		config: config.DaemonConfig{
			DNSProvisioners: []config.DNSProvisionerConfig{
				{
					Name:    "dummy",
					Config:  map[string]string{},
					Domains: []string{"bar.baz"},
				},
			},
		},
		dnsProvider: providerMock,
	}

	dbMock.EXPECT().
		FindAlias("foo", "bar.baz").
		Return(database.Alias{
			Model:  gorm.Model{ID: 42},
			Domain: "bar.baz",
			Host:   "foo",
			Value:  "127.0.0.1",
			UserID: 1,
		}, nil)

	providerMock.EXPECT().GetProvisioner("dummy", map[string]string{}).Return(provisionerMock, nil)
	provisionerMock.EXPECT().UpdateRecord("foo", "bar.baz", "8.8.8.8").Return(nil)

	dbMock.EXPECT().UpdateAlias(database.Alias{
		Model:  gorm.Model{ID: 42},
		Domain: "bar.baz",
		Host:   "foo",
		Value:  "8.8.8.8",
		UserID: uint(1),
	}).Return(database.Alias{
		Model:  gorm.Model{ID: 42},
		Domain: "bar.baz",
		Host:   "foo",
		Value:  "8.8.8.8",
		UserID: 1,
	}, nil)

	a, err := d.UpdateAlias(proto.UserContext{UserID: 1}, proto.AliasDto{Domain: "foo.bar.baz", Value: "8.8.8.8"})
	if err != nil {
		t.Error(err)
	}

	if a.Domain != "foo.bar.baz" || a.Value != "8.8.8.8" {
		t.Error("Alias not updated")
	}
}

func TestDaemon_DeleteAlias(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	dbMock := database.NewMockConnection(mockCtrl)
	provisionerMock := dns.NewMockProvisioner(mockCtrl)
	providerMock := dns.NewMockProvider(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   dbMock,
		config: config.DaemonConfig{
			DNSProvisioners: []config.DNSProvisionerConfig{
				{
					Name:    "dummy",
					Config:  map[string]string{},
					Domains: []string{"creekorful.be"},
				},
			},
		},
		dnsProvider: providerMock,
	}

	providerMock.EXPECT().GetProvisioner("dummy", map[string]string{}).Return(provisionerMock, nil)
	provisionerMock.EXPECT().DeleteRecord("www", "creekorful.be").Return(nil)

	dbMock.EXPECT().DeleteAlias("www", "creekorful.be", uint(1)).Return(nil)

	if err := d.DeleteAlias(proto.UserContext{UserID: 1}, "www.creekorful.be"); err != nil {
		t.Error(err)
	}
}

func TestDaemon_GetDomains(t *testing.T) {
	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)

	d := daemon{
		logger: &logger,
		config: config.DaemonConfig{
			DNSProvisioners: []config.DNSProvisionerConfig{
				{
					Name:    "dummy",
					Config:  map[string]string{},
					Domains: []string{"bar.baz"},
				},
				{
					Name:    "example",
					Config:  map[string]string{},
					Domains: []string{"example.org", "dydns.org"},
				},
			},
		},
	}

	domains, err := d.GetDomains(proto.UserContext{})
	if err != nil {
		t.Error(err)
	}

	if len(domains) != 3 {
		t.Error("Wrong number of domains returned")
	}

	// TODO assert on domains
}
