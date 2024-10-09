package domain

import "dbhose/pkg"

type Credential struct {
	ID        string           `json:"id"`
	Email     string           `json:"email"`
	Secret    CredentialSecret `json:"secret"`
	CreatedAt string           `json:"created_at"`
	UpdatedAt string           `json:"updated_at"`
}

type CredentialSecret struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	DBName   string `json:"dbname"`
}

func (c *Credential) Encrypt(secret string) (err error) {
	c.Secret.User, err = pkg.Encrypt(c.Secret.User, secret)
	if err != nil {
		return err
	}

	c.Secret.Password, err = pkg.Encrypt(c.Secret.Password, secret)
	if err != nil {
		return err
	}

	c.Secret.Port, err = pkg.Encrypt(c.Secret.Port, secret)
	if err != nil {
		return err
	}

	c.Secret.Host, err = pkg.Encrypt(c.Secret.Host, secret)
	if err != nil {
		return err
	}

	c.Secret.DBName, err = pkg.Encrypt(c.Secret.DBName, secret)
	if err != nil {
		return err
	}

	return
}

func (c *Credential) Decrypt(secret string) (err error) {
	c.Secret.User, err = pkg.Decrypt(c.Secret.User, secret)
	if err != nil {
		return err
	}

	c.Secret.Password, err = pkg.Decrypt(c.Secret.Password, secret)
	if err != nil {
		return err
	}

	c.Secret.Port, err = pkg.Decrypt(c.Secret.Port, secret)
	if err != nil {
		return err
	}

	c.Secret.Host, err = pkg.Decrypt(c.Secret.Host, secret)
	if err != nil {
		return err
	}

	c.Secret.DBName, err = pkg.Decrypt(c.Secret.DBName, secret)
	if err != nil {
		return err
	}

	return
}
