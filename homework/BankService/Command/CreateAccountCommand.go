package command

import (
	facade "github.com/ilyaytrewq/kpo-sb/homework/BankService/Facade"
)

type CreateAccountCommand struct {
	Facade  *facade.BankAccountFacade
	Name    string
	Balance float64
}

func (c *CreateAccountCommand) Execute() error {
	_, err := c.Facade.CreateAccount(c.Name, c.Balance)
	return err
}
