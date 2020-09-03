package database

import (
	"errors"
	"github.com/creekorful/open-dydns/internal/opendydnsd/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Email    string `gorm:"unique"`
	Password string

	Aliases []Alias
}

type Alias struct {
	gorm.Model

	// TODO split domain & host
	Domain string
	Value  string
	UserID uint // FK
}

type Connection interface {
	CreateUser(email, hashedPassword string) (User, error)
	FindUser(email string) (User, error)
	FindUserAliases(userId uint) ([]Alias, error)
	FindAlias(name string) (Alias, error)
	CreateAlias(alias Alias, userId uint) (Alias, error)
	DeleteAlias(name string, userId uint) error
	UpdateAlias(alias Alias, userId uint) (Alias, error)
}

type connection struct {
	connection *gorm.DB
}

func OpenConnection(d config.DatabaseConfig) (Connection, error) {
	// TODO support multiple provider using config
	// TODO use factory pattern
	conn, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: &nopLogger{},
	})
	if err != nil {
		return nil, err
	}

	// TODO remove? better?
	if err := conn.AutoMigrate(&Alias{}, &User{}); err != nil {
		return nil, err
	}

	// TODO remove this code
	c := &connection{connection: conn}

	// Create demo user
	_, err = c.FindUser("lunamicard@gmail.com")
	if errors.As(err, &gorm.ErrRecordNotFound) {
		if _, err := c.CreateUser("lunamicard@gmail.com", "test"); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *connection) CreateUser(email, hashedPassword string) (User, error) {
	user := User{
		Email:    email,
		Password: hashedPassword,
	}

	result := c.connection.Create(&user)
	return user, result.Error
}

func (c *connection) FindUser(email string) (User, error) {
	var user User
	result := c.connection.Where("email = ?", email).First(&user)
	return user, result.Error
}

func (c *connection) FindUserAliases(userId uint) ([]Alias, error) {
	var aliases []Alias
	err := c.connection.Model(&User{Model: gorm.Model{ID: userId}}).Association("Aliases").Find(&aliases)
	return aliases, err
}

func (c *connection) FindAlias(name string) (Alias, error) {
	var alias Alias
	result := c.connection.Where("domain = ?", name).First(&alias)
	return alias, result.Error
}

func (c *connection) CreateAlias(alias Alias, userId uint) (Alias, error) {
	err := c.connection.Model(&User{Model: gorm.Model{ID: userId}}).Association("Aliases").Append(&alias)
	return alias, err
}

func (c *connection) DeleteAlias(name string, userId uint) error {
	result := c.connection.Where("domain = ?", name).Delete(Alias{}) // TODO restrict userId
	return result.Error
}

func (c *connection) UpdateAlias(alias Alias, userId uint) (Alias, error) {
	result := c.connection.Model(&alias).Updates(Alias{ // TODO restrict userId
		Domain: alias.Domain,
		Value:  alias.Value,
	})
	return alias, result.Error
}
