package conf

import (
	"v2ray.com/core/app/admin"
)

type Account struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}
type AdminConfig struct {
	Addr        string     `json:"addr"`
	ContextPath string     `json:"contextPath"`
	PublicPath  string     `json:"publicPath"`
	Accounts    []*Account `json:"accounts"`
}

func (c *AdminConfig) Build() (*admin.Config, error) {
	if c.Addr == "" {
		return nil, newError("admin addr can't be empty.")
	}

	var accounts []*admin.Account
	if c.Accounts != nil && len(c.Accounts) > 0 {
		accounts = make([]*admin.Account, 0, 10)
		for _, account := range c.Accounts {
			accounts = append(accounts, &admin.Account{
				UserName: account.UserName,
				Password: account.Password,
			})
		}
	}

	return &admin.Config{
		Addr:        c.Addr,
		ContextPath: c.ContextPath,
		PublicPath:  c.PublicPath,
		Accounts:    accounts,
	}, nil
}
