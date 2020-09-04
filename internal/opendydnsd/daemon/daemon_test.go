package daemon

import (
	"errors"
	"github.com/creekorful/open-dydns/internal/opendydnsd/database"
	"github.com/creekorful/open-dydns/internal/proto"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"io/ioutil"
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
	mockObj := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   mockObj,
	}

	mockObj.EXPECT().
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
	mockObj := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   mockObj,
	}

	pass, err := d.hashPassword("test")
	if err != nil {
		t.Error(err)
	}

	mockObj.EXPECT().
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
	mockObj := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   mockObj,
	}

	pass, err := d.hashPassword("test")
	if err != nil {
		t.Error(err)
	}

	mockObj.EXPECT().
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
	// TODO
}

func TestDaemon_RegisterAlias_InvalidRequest(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	mockObj := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   mockObj,
	}

	_, err := d.RegisterAlias(proto.UserContext{UserID: 1}, proto.AliasDto{})
	if !errors.As(err, &ErrInvalidParameters) {
		t.Error("RegisterAlias() should have returned ErrInvalidParameters")
	}
}

func TestDaemon_RegisterAlias_AliasTaken(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	mockObj := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   mockObj,
	}

	mockObj.EXPECT().FindAlias("creekorful.de").Return(database.Alias{
		Domain: "creekorful.de",
		UserID: 12,
	}, nil)

	_, err := d.RegisterAlias(proto.UserContext{UserID: 1}, proto.AliasDto{
		Domain: "creekorful.de", Value: "127.0.0.1",
	})

	if !errors.As(err, &ErrAliasTaken) {
		t.Error("RegisterAlias() should have returned ErrAliasTaken")
	}
}

func TestDaemon_RegisterAlias_AliasAlreadyExist(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	mockObj := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   mockObj,
	}

	mockObj.EXPECT().FindAlias("creekorful.de").Return(database.Alias{
		Domain: "creekorful.de",
		UserID: 1,
	}, nil)

	_, err := d.RegisterAlias(proto.UserContext{UserID: 1}, proto.AliasDto{
		Domain: "creekorful.de", Value: "127.0.0.1",
	})

	if !errors.As(err, &ErrAliasAlreadyExist) {
		t.Error("RegisterAlias() should have returned ErrAliasAlreadyExist")
	}
}

func TestDaemon_RegisterAlias(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	mockObj := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   mockObj,
	}

	mockObj.EXPECT().
		FindAlias("creekorful.de").
		Return(database.Alias{}, gorm.ErrRecordNotFound)
	mockObj.EXPECT().
		CreateAlias(database.Alias{Domain: "creekorful.de", Value: "127.0.0.1"}, uint(1)).
		Return(database.Alias{
			Model:  gorm.Model{ID: 12},
			Domain: "creekorful.de",
			Value:  "127.0.0.1",
			UserID: 1,
		}, nil)

	r, err := d.RegisterAlias(proto.UserContext{UserID: 1}, proto.AliasDto{
		Domain: "creekorful.de", Value: "127.0.0.1",
	})

	if err != nil {
		t.Error(err)
	}

	if r.Domain != "creekorful.de" || r.Value != "127.0.0.1" {
		t.Error("Wrong alias created")
	}
}

func TestDaemon_UpdateAlias_AliasDoesNotExist(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.Output(ioutil.Discard).Level(zerolog.Disabled)
	mockObj := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   mockObj,
	}

	mockObj.EXPECT().
		FindAlias("foo.bar.baz").
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
	mockObj := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   mockObj,
	}

	mockObj.EXPECT().
		FindAlias("foo.bar.baz").
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
	mockObj := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   mockObj,
	}

	mockObj.EXPECT().
		FindAlias("foo.bar.baz").
		Return(database.Alias{
			Model:  gorm.Model{ID: 42},
			Domain: "foo.bar.baz",
			Value:  "127.0.0.1",
			UserID: 1,
		}, nil)
	mockObj.EXPECT().UpdateAlias(database.Alias{
		Model:  gorm.Model{ID: 42},
		Domain: "foo.bar.baz",
		Value:  "8.8.8.8",
		UserID: 1,
	}).Return(database.Alias{
		Model:  gorm.Model{ID: 42},
		Domain: "foo.bar.baz",
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
	mockObj := database.NewMockConnection(mockCtrl)

	d := daemon{
		logger: &logger,
		conn:   mockObj,
	}

	mockObj.EXPECT().DeleteAlias("creekorful.be", uint(1)).Return(nil)

	if err := d.DeleteAlias(proto.UserContext{UserID: 1}, "creekorful.be"); err != nil {
		t.Error(err)
	}
}
