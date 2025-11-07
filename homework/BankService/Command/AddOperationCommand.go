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
}

func (c *AddOperationCommand) Execute() error {
	_, err := c.Facade.CreateOperation(c.Type, c.AccountID, c.Amount, c.Date, c.CategoryID, c.Description)
	return err
}
