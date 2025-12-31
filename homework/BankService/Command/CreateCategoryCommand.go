package command

import (
	facade "github.com/ilyaytrewq/kpo-sb/homework/BankService/Facade"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
)

type CreateCategoryCommand struct {
	Facade    *facade.CategoryFacade
	Name      string
	Type      category.CategoryType
	CreatedID service.ObjectID
}

func (c *CreateCategoryCommand) Execute() error {
	id, err := c.Facade.CreateCategory(c.Name, c.Type)
	if err != nil {
		return err
	}
	c.CreatedID = id
	return nil
}
