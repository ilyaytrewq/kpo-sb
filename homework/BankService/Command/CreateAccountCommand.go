package command

import (
	facade "github.com/ilyaytrewq/kpo-sb/homework/BankService/Facade"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
)

type CreateAccountCommand struct {
	Facade    *facade.BankAccountFacade
	Name      string
	Balance   float64
	CreatedID service.ObjectID
}

func (c *CreateAccountCommand) Execute() error {
	id, err := c.Facade.CreateAccount(c.Name, c.Balance)
	if err != nil {
		return err
	}
	c.CreatedID = id
	return nil
}
