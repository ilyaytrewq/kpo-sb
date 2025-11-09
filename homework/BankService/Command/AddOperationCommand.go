package command

import (
	"time"

	facade "github.com/ilyaytrewq/kpo-sb/homework/BankService/Facade"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
)

type AddOperationCommand struct {
	Facade      *facade.OperationFacade
	Type        operation.OperationType
	AccountID   service.ObjectID
	Amount      float64
	Date        time.Time
	CategoryID  service.ObjectID
	Description string
	CreatedID   service.ObjectID
}

func (c *AddOperationCommand) Execute() error {
	id, err := c.Facade.CreateOperation(c.Type, c.AccountID, c.Amount, c.Date, c.CategoryID, c.Description)
	if err != nil {
		return err
	}
	c.CreatedID = id
	return nil
}
